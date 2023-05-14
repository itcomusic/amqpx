package amqpx

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type PublishInterceptor func(PublishFunc) PublishFunc

// PublishFunc is func used for publish message.
type PublishFunc func(context.Context, *PublishingRequest) error

// A PublishingRequest represents a request to publish a message.
type PublishingRequest struct {
	amqp091.Publishing
	opts publishOptions
}

// A Publishing represents message sending to the server.
type Publishing[T any] struct {
	msg *T
	req *PublishingRequest
}

// NewPublishing creates new publishing.
func NewPublishing[T any](v *T) *Publishing[T] {
	return &Publishing[T]{
		msg: v,
		req: &PublishingRequest{
			Publishing: amqp091.Publishing{Headers: make(amqp091.Table)},
		},
	}
}

// PersistentMode sets delivery mode as persistent.
func (m *Publishing[T]) PersistentMode() *Publishing[T] {
	m.req.DeliveryMode = Persistent
	return m
}

// SetPriority sets priority.
func (m *Publishing[T]) SetPriority(priority uint8) *Publishing[T] {
	m.req.Priority = priority
	return m
}

// SetCorrelationID sets correlation id.
func (m *Publishing[T]) SetCorrelationID(id string) *Publishing[T] {
	m.req.CorrelationId = id
	return m
}

// SetReplyTo sets reply to.
func (m *Publishing[T]) SetReplyTo(replyTo string) *Publishing[T] {
	m.req.ReplyTo = replyTo
	return m
}

// SetExpiration sets expiration.
func (m *Publishing[T]) SetExpiration(expiration string) *Publishing[T] {
	m.req.Expiration = expiration
	return m
}

// SetMessageID sets message id.
func (m *Publishing[T]) SetMessageID(id string) *Publishing[T] {
	m.req.MessageId = id
	return m
}

// SetTimestamp sets timestamp.
func (m *Publishing[T]) SetTimestamp(timestamp time.Time) *Publishing[T] {
	m.req.Timestamp = timestamp
	return m
}

// SetType sets message type.
func (m *Publishing[T]) SetType(typ string) *Publishing[T] {
	m.req.Type = typ
	return m
}

// SetUserID sets user id.
func (m *Publishing[T]) SetUserID(id string) *Publishing[T] {
	m.req.UserId = id
	return m
}

// SetAppID sets application id.
func (m *Publishing[T]) SetAppID(id string) *Publishing[T] {
	m.req.AppId = id
	return m
}

// A Publisher represents client for sending the messages.
type Publisher[T any] struct {
	amqpChannel      atomic.Pointer[Channel]
	conn             func() Connection
	notifyAMQPClose  chan *amqp091.Error
	notifyAMQPCancel chan string
	exchange         string
	confirm          bool
	publishExec      PublishFunc

	marshaler      Marshaler
	publishOptions publishOptions
	log            LogFunc
	done           context.Context
	cancel         context.CancelFunc
	err            error
}

// NewPublisher creates a publisher.
func NewPublisher[T any](client *Client, exchange string, opts ...PublisherOption) *Publisher[T] {
	opt := &publisherOptions{
		marshaler:   client.marshaler,
		interceptor: client.publishInterceptor,
		publish: publishOptions{
			ctx: context.Background(),
		},
	}

	if _, ok := any(new(T)).(*[]byte); ok {
		opt.marshaler = defaultBytesMarshaler
	}
	for _, o := range opts {
		o(opt)
	}

	if err := opt.validate(); err != nil {
		return &Publisher[T]{err: err}
	}

	pub := &Publisher[T]{
		conn: client.conn,
		// default close channels to reconnect
		notifyAMQPClose: func() chan *amqp091.Error {
			ch := make(chan *amqp091.Error)
			close(ch)
			return ch
		}(),
		notifyAMQPCancel: func() chan string {
			ch := make(chan string)
			close(ch)
			return ch
		}(),
		exchange:       exchange,
		confirm:        opt.confirm,
		publishOptions: opt.publish,
		marshaler:      opt.marshaler,
		log:            client.logger,
	}
	pub.done, pub.cancel = context.WithCancel(client.done)

	// wrap the end fn with the interceptor chain
	pub.publishExec = pub.publish
	if len(opt.interceptor) != 0 {
		pub.publishExec = opt.interceptor[len(opt.interceptor)-1](pub.publishExec)
		for i := len(opt.interceptor) - 2; i >= 0; i-- {
			pub.publishExec = opt.interceptor[i](pub.publishExec)
		}
	}

	_ = pub.initChannel()
	go pub.serve()
	return pub
}

// Publish publishes the message.
func (p *Publisher[T]) Publish(m *Publishing[T], opts ...PublishOption) error {
	if err := p.err; err != nil {
		return p.newPublishError(m.req.opts.key, err)
	}

	m.req.opts = p.publishOptions
	for _, v := range opts {
		v(&m.req.opts)
	}

	if err := m.req.opts.validate(); err != nil {
		return p.newPublishError(m.req.opts.key, err)
	}

	b, err := p.marshaler.Marshal(m.msg)
	if err != nil {
		return p.newPublishError(m.req.opts.key, err)
	}
	m.req.Body = b
	m.req.ContentType = p.marshaler.ContentType()

	return p.publishExec(m.req.opts.ctx, m.req)
}

func (p *Publisher[T]) publish(ctx context.Context, m *PublishingRequest) error {
	channel := p.amqpChannel.Load()
	if channel == nil {
		return p.newPublishError(m.opts.key, errChannelClosed)
	}

	confirm, err := (*channel).PublishWithDeferredConfirmWithContext(ctx, p.exchange, m.opts.key, m.opts.mandatory, m.opts.immediate, m.Publishing)
	if err != nil {
		return p.newPublishError(m.opts.key, err)
	}

	if confirm != nil {
		ok, err := confirm.WaitContext(ctx)
		if err != nil {
			return p.newPublishError(m.opts.key, fmt.Errorf("%s: %w", errPublishConfirm, err))
		}

		if !ok {
			return p.newPublishError(m.opts.key, errPublishConfirm)
		}
	}
	return nil
}

// Close closes publisher.
func (p *Publisher[T]) Close() {
	p.cancel()
}

func (p *Publisher[T]) setChannel(channel Channel) {
	if c := p.amqpChannel.Load(); c != nil {
		(*c).Close()
	}
	p.amqpChannel.Store(&channel)
}

func (p *Publisher[T]) initChannel() error {
	conn := p.conn()
	if conn.IsClosed() {
		return errConnClosed
	}

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("create channel: %w", err)
	}

	if p.confirm {
		if err := channel.Confirm(false); err != nil {
			channel.Close()
			return fmt.Errorf("confirm mode: %w", err)
		}
	}

	p.setChannel(channel)
	p.notifyAMQPClose = channel.NotifyClose(make(chan *amqp091.Error, 1))
	p.notifyAMQPCancel = channel.NotifyCancel(make(chan string, 1))
	go p.notifyReturn(channel)
	return nil
}

func (p *Publisher[T]) serve() {
	defer func() {
		if channel := p.amqpChannel.Load(); channel != nil {
			(*channel).Close()
		}
	}()

	for {
		select {
		case <-p.done.Done():
			return

		case <-p.notifyAMQPClose:
		case <-p.notifyAMQPCancel:
		}

		if exit := p.makeConnect(); exit {
			return
		}
	}
}

func (p *Publisher[T]) makeConnect() (exit bool) {
	for {
		var err error
		if err = p.initChannel(); err == nil {
			return false
		}

		if !errors.Is(err, errConnClosed) {
			p.log("[ERROR] exchange %q routing-key %q: %s", p.exchange, p.publishOptions.key, err)
		}

		select {
		case <-p.done.Done():
			return true

		case <-time.After(defaultReconnectDelay):
		}
	}
}

func (p *Publisher[T]) notifyReturn(channel Channel) {
	for v := range channel.NotifyReturn(make(chan amqp091.Return, 1)) {
		p.log("[ERROR] exchange %q routing-key %q undeliverable message desc %q \"%d\"", v.Exchange, v.RoutingKey, v.ReplyText, v.ReplyCode)
	}
}

func (p *Publisher[T]) newPublishError(routingKey string, err error) error {
	return fmt.Errorf("amqpx: exchange %q routing-key %q: %s", p.exchange, routingKey, err)
}

package amqpx

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type PublishHook func(PublisherFunc) PublisherFunc

// PublisherFunc is func used for publish message.
type PublisherFunc func(*Publishing) error

// A Publishing represents message sending to the server.
type Publishing struct {
	amqp091.Publishing
	ctx  context.Context
	err  error
	opts publishOptions
}

func (m *Publishing) amqp091() amqp091.Publishing {
	return m.Publishing
}

// PersistentMode sets delivery mode as persistent.
func (m *Publishing) PersistentMode() *Publishing {
	m.DeliveryMode = Persistent
	return m
}

// Context returns context.
func (m *Publishing) Context() context.Context {
	return m.ctx
}

func (m *Publishing) WithContext(ctx context.Context) {
	m.ctx = ctx
}

// A Publisher represents client for sending the messages.
type Publisher[T any] struct {
	amqpChannel      atomic.Value // Channel
	conn             func() Connection
	notifyAMQPClose  chan *amqp091.Error
	notifyAMQPCancel chan string
	exchange         string
	confirm          bool
	fn               PublisherFunc

	marshaler      Marshaler
	publishOptions publishOptions
	logFunc        LogFunc
	done           context.Context
	cancel         context.CancelFunc
	err            error
}

// NewPublisher creates a publisher.
func NewPublisher[T any](c *Client, exchange string, opts ...PublisherOption) *Publisher[T] {
	opt := &publisherOptions{
		publish:   publishOptions{},
		marshaler: c.marshaler,
		hook:      c.publishHook,
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
		conn: c.conn,
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
		logFunc:        c.logger,
	}
	pub.done, pub.cancel = context.WithCancel(c.done)

	// wrap the end fn with the hook chain
	pub.fn = pub.publish
	if len(opt.hook) != 0 {
		pub.fn = opt.hook[len(opt.hook)-1](pub.fn)
		for i := len(opt.hook) - 2; i >= 0; i-- {
			pub.fn = opt.hook[i](pub.fn)
		}
	}

	_ = pub.initChannel()
	go pub.serve()
	return pub
}

// NewPublishing creates new publishing.
func (p *Publisher[T]) NewPublishing(v T) *Publishing {
	if p.err != nil {
		return &Publishing{}
	}

	b, err := p.marshaler.Marshal(v)
	if err != nil {
		return &Publishing{err: err}
	}
	return &Publishing{Publishing: amqp091.Publishing{Body: b, ContentType: p.marshaler.ContentType(), Headers: make(amqp091.Table)}}
}

// Publish publishes the message.
func (p *Publisher[T]) Publish(m *Publishing, opts ...PublishOption) error {
	if p.err != nil {
		return p.err
	}

	m.opts = p.publishOptions
	for _, v := range opts {
		v(&m.opts)
	}
	return p.fn(m)
}

func (p *Publisher[T]) publish(m *Publishing) error {
	if m.err != nil {
		return m.err
	}

	channel, ok := p.amqpChannel.Load().(Channel)
	if !ok {
		return ErrChannelClosed
	}

	confirm, err := channel.PublishWithDeferredConfirm(p.exchange, m.opts.key, m.opts.mandatory, m.opts.mandatory, m.amqp091())
	if err != nil {
		return fmt.Errorf("amqpx: publish: %w", err)
	}

	if confirm != nil && !confirm.Wait() {
		return ErrPublishConfirm
	}
	return nil
}

// Close closes publisher.
func (p *Publisher[T]) Close() {
	p.cancel()
}

func (p *Publisher[T]) setChannel(channel Channel) {
	if c, ok := p.amqpChannel.Load().(Channel); ok {
		c.Close()
	}
	p.amqpChannel.Store(channel)
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
			return fmt.Errorf("channel confirm mode: %w", err)
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
		if channel, ok := p.amqpChannel.Load().(Channel); ok {
			channel.Close()
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
			p.logFunc("[ERROR] init channel: %s", err)
		}

		select {
		case <-p.done.Done():
			return true

		case <-time.After(reconnectDelay):
		}
	}
}

func (p *Publisher[T]) notifyReturn(channel Channel) {
	for v := range channel.NotifyReturn(make(chan amqp091.Return, 1)) {
		p.logFunc("[ERROR] \"%s\" \"%s\" undeliverable message desc: %s (%d)", v.Exchange, v.RoutingKey, v.ReplyText, v.ReplyCode)
	}
}

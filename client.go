package amqpx

import (
	"context"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/semaphore"
)

const (
	defaultConnectTimeout = time.Second * 5
)

// A Client represents connection to rabbitmq.
type Client struct {
	dialer      dialer
	mx          sync.RWMutex
	amqpConn    Connection
	marshaler   Marshaler
	unmarshaler map[string]Unmarshaler

	notifyClose chan *amqp091.Error
	wg          *sync.WaitGroup
	done        context.Context
	cancel      context.CancelFunc

	logger      LogFunc
	consumeHook []ConsumeHook
	publishHook []PublishHook
}

// Connect creates a connection.
func Connect(opts ...ClientOption) (*Client, error) {
	opt := newClientOptions()
	for _, o := range opts {
		o(&opt)
	}

	if err := opt.validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultConnectTimeout)
	defer cancel()

	amqpConn, err := opt.dialer.Dial(ctx)
	if err != nil {
		return nil, err
	}

	conn := &Client{
		dialer:      opt.dialer,
		amqpConn:    amqpConn,
		marshaler:   opt.marshaler,
		unmarshaler: opt.unmarshaler,
		notifyClose: amqpConn.NotifyClose(make(chan *amqp091.Error, 1)),
		wg:          &sync.WaitGroup{},
		logger:      opt.logger,
	}
	conn.done, conn.cancel = context.WithCancel(context.Background())
	go conn.loop()

	return conn, nil
}

// NewConsumer creates a consumer.
func (c *Client) NewConsumer(queue string, fn Consume, opts ...ConsumerOption) error {
	opt := consumerOptions{
		hook:        c.consumeHook,
		unmarshaler: c.unmarshaler,
	}
	for _, o := range opts {
		o(&opt)
	}

	if err := opt.validate(fn); err != nil {
		return ConsumerError{Queue: queue, Tag: opt.tag, Message: err.Error()}
	}

	fn.init(opt.unmarshaler)
	cons := &consumer{
		conn:    c.conn,
		queue:   queue,
		tag:     opt.tag,
		opts:    opt.channel,
		logFunc: c.logger,
		limit:   semaphore.NewWeighted(int64(opt.concurrency)),
		wg:      c.wg,
		fn:      fn,
		done:    c.done,
	}

	// wrap the end fn with the hook chain
	if len(opt.hook) != 0 {
		cons.fn = opt.hook[len(opt.hook)-1](cons.fn)
		for i := len(opt.hook) - 2; i >= 0; i-- {
			cons.fn = opt.hook[i](cons.fn)
		}
	}

	if err := cons.initChannel(); err != nil {
		return cons.newConsumerError(err)
	}
	go cons.serve()
	return nil
}

// Close closes Connection.
// Waits all consumers.
func (c *Client) Close() {
	c.cancel()
	c.wg.Wait()

	c.mx.Lock()
	defer c.mx.Unlock()
	c.amqpConn.Close()
}

func (c *Client) setConn(conn Connection) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.amqpConn.Close()
	c.amqpConn = conn
	c.notifyClose = c.amqpConn.NotifyClose(make(chan *amqp091.Error, 1))
}

func (c *Client) conn() Connection {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.amqpConn
}

func (c *Client) loop() {
	for {
		select {
		case <-c.done.Done():
			return

		case <-c.notifyClose:
			conn, err := c.dialer.Dial(c.done)
			if err != nil {
				return
			}
			c.setConn(conn)
		}
	}
}

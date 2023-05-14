package amqpx

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/semaphore"
)

//go:generate ./bin/moq -rm -out acknowledger_moq_test.go . Acknowledger

type Acknowledger interface {
	Ack(tag uint64, multiple bool) error
	Nack(tag uint64, multiple bool, requeue bool) error
	Reject(tag uint64, requeue bool) error
}

type ConsumeInterceptor func(ConsumeFunc) ConsumeFunc

type ConsumeFunc func(ctx context.Context, req *DeliveryRequest) Action

type HandlerValue interface {
	serve(ctx context.Context, req *DeliveryRequest) Action
	init(map[string]Unmarshaler)
}

// handleValue represents consume message unmarshales bytes into
// the appropriate struct based on the signature of the func.
type handleValue[T any] struct {
	fn          func(context.Context, *Delivery[T]) Action
	unmarshaler map[string]Unmarshaler
	bytesMsg    bool
}

// D represents handler of consume amqpx.Delivery[T].
func D[T any](fn func(ctx context.Context, d *Delivery[T]) Action) *handleValue[T] {
	if _, ok := any(new(T)).(*[]byte); ok {
		return &handleValue[T]{fn: fn, bytesMsg: true}
	}

	return &handleValue[T]{fn: fn}
}

func (v *handleValue[T]) init(m map[string]Unmarshaler) {
	v.unmarshaler = m
}

func (v *handleValue[T]) serve(ctx context.Context, req *DeliveryRequest) Action {
	if v.bytesMsg {
		return v.fn(ctx, &Delivery[T]{Msg: any(&req.Body).(*T), Req: req})
	}

	u, ok := v.unmarshaler[req.ContentType]
	if !ok {
		req.log("[ERROR] %s: %s", req.info(), errUnmarshalerNotFound)
		return Reject
	}

	value := new(T)
	if err := u.Unmarshal(req.Body, value); err != nil {
		req.log("[ERROR] %s: %s", req.info(), fmt.Errorf("has an error trying to unmarshal: %w", err))
		return Reject
	}

	req.Body = nil
	return v.fn(ctx, newDelivery(value, req))
}

type consumer struct {
	conn             func() Connection
	channel          Channel
	notifyAMQPClose  chan *amqp091.Error
	notifyAMQPCancel chan string
	delivery         channelDelivery
	queue            string
	tag              string
	opts             channelOptions

	limit *semaphore.Weighted
	wg    *sync.WaitGroup
	fn    ConsumeFunc

	log  LogFunc
	done context.Context
}

func (c *consumer) initChannel() error {
	conn := c.conn()
	if conn.IsClosed() {
		return errConnClosed
	}

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("create channel: %w", err)
	}

	// declare
	if c.opts.queueDeclare != nil {
		if _, err := channel.QueueDeclare(c.queue,
			c.opts.queueDeclare.Durable,
			c.opts.queueDeclare.AutoDelete,
			c.opts.queueDeclare.Exclusive,
			c.opts.queueDeclare.NoWait,
			c.opts.queueDeclare.Args); err != nil {
			return fmt.Errorf("declare queue: %w", err)
		}
	}

	if c.opts.exchangeDeclare != nil {
		if err := channel.ExchangeDeclare(c.opts.exchangeDeclare.Name,
			c.opts.exchangeDeclare.Type,
			c.opts.exchangeDeclare.Durable,
			c.opts.exchangeDeclare.AutoDelete,
			c.opts.exchangeDeclare.Internal,
			c.opts.exchangeDeclare.NoWait,
			c.opts.exchangeDeclare.Args); err != nil {
			return fmt.Errorf("declare exchange: %w", err)
		}
	}

	if c.opts.queueBind != nil {
		for _, k := range c.opts.queueBind.RoutingKey {
			if err := channel.QueueBind(c.queue, k,
				c.opts.queueBind.Exchange,
				c.opts.queueBind.NoWait,
				c.opts.queueBind.Args,
			); err != nil {
				return fmt.Errorf("bind queue: %w", err)
			}
		}
	}

	// prefetch
	if err := channel.Qos(c.opts.prefetchCount, 0, false); err != nil {
		return fmt.Errorf("qos: %w", err)
	}

	// close old channel
	if c.channel != nil {
		c.channel.Close()
	}
	c.channel = channel

	delivery, err := c.channel.Consume(c.queue, c.tag, c.opts.autoAck, c.opts.exclusive, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	c.notifyAMQPClose = c.channel.NotifyClose(make(chan *amqp091.Error, 1))
	c.notifyAMQPCancel = c.channel.NotifyCancel(make(chan string, 1))
	c.delivery.init(c.done, delivery)
	return nil
}

func (c *consumer) serve() {
	defer c.close()

	for {
		select {
		case <-c.done.Done():
			return

		case <-c.notifyAMQPClose:
		case <-c.notifyAMQPCancel:

		case d, ok := <-c.delivery.channel:
			if !ok {
				break
			}

			c.wg.Add(1)
			if err := c.limit.Acquire(c.done, 1); err != nil {
				return
			}

			go c.handleDelivery(c.delivery.ctx, &d)
			continue
		}

		if exit := c.makeConnect(); exit {
			return
		}
	}
}

func (c *consumer) makeConnect() (exit bool) {
	c.delivery.cancel()

	for {
		var err error
		if err = c.initChannel(); err == nil {
			return false
		}

		if !errors.Is(err, errConnClosed) {
			c.log("[ERROR] queue %q consumer-tag %q: %s", c.queue, c.tag, err)
		}

		select {
		case <-c.done.Done():
			return true

		case <-time.After(defaultReconnectDelay):
		}
	}
}

func (c *consumer) handleDelivery(ctx context.Context, d *amqp091.Delivery) {
	defer c.wg.Done()
	defer c.limit.Release(1)

	delivery := newDeliveryRequest(d, c.log)
	if status := c.fn(ctx, delivery); !c.opts.autoAck {
		if err := delivery.setStatus(status); err != nil {
			c.log("[ERROR] queue %q consumer-tag %q: %s", c.queue, c.tag, err)
		}
	}
}

func (c *consumer) close() {
	c.delivery.cancel()
	c.channel.Close()
}

type channelDelivery struct {
	ctx     context.Context
	cancel  context.CancelFunc
	channel <-chan amqp091.Delivery
}

func (c *channelDelivery) init(ctx context.Context, ch <-chan amqp091.Delivery) {
	if c.cancel != nil {
		c.cancel()
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.channel = ch
}

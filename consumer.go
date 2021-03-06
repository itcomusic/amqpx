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

type ConsumeHook func(Consume) Consume

type Consume interface {
	init(map[string]Unmarshaler)
	Serve(*Delivery) Action
}

// D represents handler of consume amqpx.Delivery.
type D func(d *Delivery) Action

func (fn D) init(_ map[string]Unmarshaler) {}

func (fn D) Serve(d *Delivery) Action {
	return fn(d)
}

// HandleValue represents consume message unmarshales bytes into
// the appropriate struct based on the signature of the func.
type HandleValue[T any] struct {
	fn          func(context.Context, *T) Action
	unmarshaler map[string]Unmarshaler
	pool        *pool[T]
}

// T returns handler of consume specific message type.
func T[T any](fn func(ctx context.Context, m *T) Action, opts ...PoolOptions[T]) *HandleValue[T] {
	pool := &pool[T]{}
	for _, o := range opts {
		o(pool)
	}
	return &HandleValue[T]{fn: fn, pool: pool}
}

func (v *HandleValue[T]) init(m map[string]Unmarshaler) {
	v.unmarshaler = m
}

func (v *HandleValue[T]) Serve(d *Delivery) Action {
	u, ok := v.unmarshaler[d.ContentType]
	if !ok {
		d.logFunc(DeliveryError{
			Exchange:   d.Exchange,
			RoutingKey: d.RoutingKey,
			Message:    fmt.Sprintf("content-type \"%s\" of the unmarshal not found", d.ContentType),
		})
		return Reject
	}

	value := v.pool.Get()
	defer v.pool.Put(value)

	if err := u.Unmarshal(d.Body, value); err != nil {
		d.logFunc(DeliveryError{
			Exchange:   d.Exchange,
			RoutingKey: d.RoutingKey,
			Message:    fmt.Sprintf("has an error trying to unmarshal: %s", err),
		})
		return Reject
	}
	return v.fn(toContext(d), value)
}

type consumer struct {
	conn             func() Connection
	channel          Channel
	notifyAMQPClose  chan *amqp091.Error
	notifyAMQPCancel chan string
	delivery         <-chan amqp091.Delivery
	queue            string
	tag              string
	opts             channelOptions

	limit *semaphore.Weighted
	wg    *sync.WaitGroup
	fn    Consume

	logFunc LogFunc
	done    context.Context
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

	d, err := c.channel.Consume(c.queue, c.tag, c.opts.autoAck, c.opts.exclusive, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	c.notifyAMQPClose = c.channel.NotifyClose(make(chan *amqp091.Error, 1))
	c.notifyAMQPCancel = c.channel.NotifyCancel(make(chan string, 1))
	c.delivery = d
	return nil
}

func (c *consumer) serve() {
	defer func() {
		if c.channel != nil {
			c.channel.Close()
		}
	}()

	for {
		select {
		case <-c.done.Done():
			return

		case <-c.notifyAMQPClose:
		case <-c.notifyAMQPCancel:

		case d, ok := <-c.delivery:
			if !ok {
				break
			}

			c.wg.Add(1)
			if err := c.limit.Acquire(c.done, 1); err != nil {
				c.wg.Done()
				return
			}

			go c.handleDelivery(&d)
			continue
		}

		if exit := c.makeConnect(); exit {
			return
		}
	}
}

func (c *consumer) makeConnect() (exit bool) {
	for {
		var err error
		if err = c.initChannel(); err == nil {
			return false
		}

		if !errors.Is(err, errConnClosed) {
			c.logFunc(c.newConsumerError(err))
		}

		select {
		case <-c.done.Done():
			return true

		case <-time.After(reconnectDelay):
		}
	}
}

func (c *consumer) handleDelivery(d *amqp091.Delivery) {
	defer c.wg.Done()
	defer c.limit.Release(1)

	delivery := newDelivery(d, c.logFunc)
	if status := c.fn.Serve(delivery); !c.opts.autoAck {
		if err := delivery.setStatus(status); err != nil {
			c.logFunc(c.newConsumerError(err))
		}
	}
}

func (c *consumer) newConsumerError(err error) ConsumerError {
	return ConsumerError{
		Queue:   c.queue,
		Tag:     c.tag,
		Message: err.Error(),
	}
}

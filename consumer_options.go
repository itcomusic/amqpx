package amqpx

import (
	"runtime"
)

var defaultLimitConcurrency = runtime.GOMAXPROCS(0)

// ConsumerOption is used to configure a consumer.
type ConsumerOption func(*consumerOptions)

type consumerOptions struct {
	tag         string
	channel     channelOptions
	concurrency int
	hook        []ConsumeHook
	unmarshaler map[string]Unmarshaler
}

type channelOptions struct {
	autoAck         bool
	exclusive       bool
	prefetchCount   int
	queueDeclare    *QueueDeclare
	exchangeDeclare *ExchangeDeclare
	queueBind       *QueueBind
}

func (c *consumerOptions) validate(fn HandlerValue) error {
	if fn == nil {
		return errFuncNil
	}

	if _, ok := fn.(*handleValue[[]byte]); !ok && len(c.unmarshaler) == 0 {
		return errUnmarshalerNotFound
	}

	if !c.channel.autoAck {
		c.concurrency = c.channel.prefetchCount
	}

	if c.concurrency == 0 {
		c.concurrency = defaultLimitConcurrency
	}
	return nil
}

// ConsumerTag sets tag.
func ConsumerTag(tag string) ConsumerOption {
	return func(o *consumerOptions) {
		o.tag = tag
	}
}

// SetAutoAckMode sets auto ack mode.
// The default is false.
func SetAutoAckMode() ConsumerOption {
	return func(o *consumerOptions) {
		o.channel.autoAck = true
	}
}

// SetExclusive sets exclusive.
// The default is false.
//
// When exclusive is true, the server will ensure that this is the sole consumer
// from this queue. When exclusive is false, the server will fairly distribute
// deliveries across multiple consumers.
func SetExclusive(b bool) ConsumerOption {
	return func(o *consumerOptions) {
		o.channel.exclusive = b
	}
}

// SetPrefetchCount sets prefetch count.
// prefetchCount controls how many messages the server will try to keep on
// the network for consumers before receiving delivery acks. The intent of prefetchCount is
// to make sure the network buffers stay full between the server and client.
//
// With a prefetch count greater than zero, the server will deliver that many
// messages to consumers before acknowledgments are received. The server ignores
// this option when consumers are started with AutoAck=false because no acknowledgments
// are expected or sent.
func SetPrefetchCount(i int) ConsumerOption {
	return func(o *consumerOptions) {
		o.channel.prefetchCount = i
	}
}

// SetConcurrency sets limit the number of goroutines for every delivered message.
// The default is runtime.GOMAXPROCS(0).
// The consumer ignores this option when prefetch count greater than zero with AutoAck=false.
func SetConcurrency(i int) ConsumerOption {
	return func(o *consumerOptions) {
		if i > 0 {
			o.concurrency = i
		}
	}
}

// SetConsumeHook sets consume hook.
func SetConsumeHook(h ...ConsumeHook) ConsumerOption {
	return func(o *consumerOptions) {
		o.hook = append(o.hook, h...)
	}
}

// SetUnmarshaler sets unmarshaler.
func SetUnmarshaler(m Unmarshaler) ConsumerOption {
	return func(o *consumerOptions) {
		if m != nil {
			o.unmarshaler = map[string]Unmarshaler{m.ContentType(): m}
		}
	}
}

// DeclareQueue sets queue declare.
func DeclareQueue(q QueueDeclare) ConsumerOption {
	return func(o *consumerOptions) {
		o.channel.queueDeclare = &q
	}
}

// DeclareExchange sets exchange declare.
func DeclareExchange(e ExchangeDeclare) ConsumerOption {
	return func(o *consumerOptions) {
		o.channel.exchangeDeclare = &e
	}
}

// BindQueue sets queue bind.
func BindQueue(q QueueBind) ConsumerOption {
	return func(o *consumerOptions) {
		o.channel.queueBind = &q
	}
}

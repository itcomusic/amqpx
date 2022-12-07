package amqpx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumerOption(t *testing.T) {
	t.Parallel()

	queueDeclare := QueueDeclare{
		Durable:    true,
		AutoDelete: true,
		Exclusive:  true,
		NoWait:     true,
		Args:       nil,
	}

	exchangeDeclare := ExchangeDeclare{
		Name:       "exchange_name_value",
		Type:       "exchange_type_value",
		Durable:    true,
		AutoDelete: true,
		Internal:   true,
		NoWait:     true,
		Args:       nil,
	}

	queueBind := QueueBind{
		Exchange:   "exchange_name_value",
		RoutingKey: []string{"routing_key_value"},
		NoWait:     true,
		Args:       nil,
	}

	got := consumerOptions{}
	for _, o := range []ConsumerOption{
		ConsumerTag("consumer_tag_value"),
		SetAutoAckMode(),
		SetExclusive(true),
		SetPrefetchCount(2),
		SetConcurrency(3),
		SetUnmarshaler(testUnmarshaler),
		DeclareQueue(queueDeclare),
		DeclareExchange(exchangeDeclare),
		BindQueue(queueBind),
	} {
		o(&got)
	}

	want := consumerOptions{
		tag: "consumer_tag_value",
		channel: channelOptions{
			autoAck:         true,
			exclusive:       true,
			prefetchCount:   2,
			queueDeclare:    &queueDeclare,
			exchangeDeclare: &exchangeDeclare,
			queueBind:       &queueBind,
		},
		concurrency: 3,
		hook:        nil,
		unmarshaler: map[string]Unmarshaler{testUnmarshaler.ContentType(): testUnmarshaler},
	}
	assert.Equal(t, want, got)
}

func TestConsumerOption_Validate(t *testing.T) {
	t.Parallel()

	fn := D(func(ctx context.Context, d *Delivery[[]byte]) Action { return Ack })
	t.Run("equals prefetch count", func(t *testing.T) {
		t.Parallel()

		got := consumerOptions{channel: channelOptions{prefetchCount: 2}, unmarshaler: map[string]Unmarshaler{"any": nil}}
		require.NoError(t, got.validate(fn))
		assert.Equal(t, 2, got.concurrency)
	})

	t.Run("auto-ack mode", func(t *testing.T) {
		t.Parallel()

		got := consumerOptions{channel: channelOptions{autoAck: true}, concurrency: 2, unmarshaler: map[string]Unmarshaler{"any": nil}}
		require.NoError(t, got.validate(fn))
		assert.Equal(t, 2, got.concurrency)
	})

	t.Run("default concurrency", func(t *testing.T) {
		t.Parallel()

		got := consumerOptions{unmarshaler: map[string]Unmarshaler{"any": nil}}
		require.NoError(t, got.validate(fn))
		assert.Equal(t, defaultLimitConcurrency, got.concurrency)
	})

	t.Run("ignore unmarshaler error", func(t *testing.T) {
		t.Parallel()

		got := (&consumerOptions{}).validate(fn)
		assert.Nil(t, got)
	})

	t.Run("func nil", func(t *testing.T) {
		t.Parallel()

		got := (&consumerOptions{}).validate(nil)
		assert.ErrorIs(t, got, errFuncNil)
	})
}

package amqpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumerOptions_Validate(t *testing.T) {
	t.Parallel()

	t.Run("equals prefetch count", func(t *testing.T) {
		t.Parallel()

		got := consumerOptions{channel: channelOptions{prefetchCount: 2}, unmarshaler: map[string]Unmarshaler{"any": nil}}
		require.NoError(t, got.validate())
		assert.Equal(t, 2, got.concurrency)
	})

	t.Run("auto-ack mode", func(t *testing.T) {
		t.Parallel()

		got := consumerOptions{channel: channelOptions{autoAck: true}, concurrency: 2, unmarshaler: map[string]Unmarshaler{"any": nil}}
		require.NoError(t, got.validate())
		assert.Equal(t, 2, got.concurrency)
	})

	t.Run("default concurrency", func(t *testing.T) {
		t.Parallel()

		got := consumerOptions{unmarshaler: map[string]Unmarshaler{"any": nil}}
		require.NoError(t, got.validate())
		assert.Equal(t, defaultLimitConcurrency, got.concurrency)
	})
}

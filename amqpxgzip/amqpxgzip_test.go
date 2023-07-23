package amqpxgzip

import (
	"context"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/itcomusic/amqpx"
)

var testGZIPBody = []byte{0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0x4a, 0xcf, 0x2f, 0xc8, 0x48, 0x2d, 0x2, 0x4, 0x0, 0x0, 0xff, 0xff, 0x2b, 0xcb, 0x2b, 0x34, 0x6, 0x0, 0x0, 0x0} // "gopher" encoded with gzip

func TestInterceptor(t *testing.T) {
	t.Parallel()

	t.Run("consume", func(t *testing.T) {
		t.Parallel()

		fn := NewInterceptor().WrapConsume(func(ctx context.Context, req *amqpx.DeliveryRequest) amqpx.Action {
			return amqpx.Ack
		})

		req := (&amqpx.DeliveryRequest{}).NewFrom(&amqp091.Delivery{ContentEncoding: headerGZIP, Body: testGZIPBody})
		status := fn(context.Background(), req)
		assert.Equal(t, amqpx.Ack, status)
		assert.Equal(t, []byte("gopher"), req.Body())
	})

	t.Run("publish", func(t *testing.T) {
		t.Parallel()

		fn := NewInterceptor().WrapPublish(func(ctx context.Context, p *amqpx.PublishingRequest) error {
			assert.Equal(t, testGZIPBody, p.Body)
			return nil
		})

		req := &amqpx.PublishingRequest{Publishing: amqp091.Publishing{Body: []byte("gopher")}}
		require.NoError(t, fn(context.Background(), req))
		assert.Equal(t, testGZIPBody, req.Body)
	})
}

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

func TestConsumer(t *testing.T) {
	t.Parallel()

	fn := Consumer()(func(ctx context.Context, req *amqpx.DeliveryRequest) amqpx.Action {
		return amqpx.Ack
	})

	d := &amqpx.DeliveryRequest{ContentEncoding: headerGZIP, Body: testGZIPBody}
	status := fn(context.Background(), d)
	assert.Equal(t, amqpx.Ack, status)
	assert.Equal(t, []byte("gopher"), d.Body)
}

func TestPublisher(t *testing.T) {
	t.Parallel()

	fn := Publisher()(func(ctx context.Context, p *amqpx.PublishRequest) error {
		assert.Equal(t, testGZIPBody, p.Body)
		return nil
	})

	p := &amqpx.PublishRequest{Publishing: amqp091.Publishing{Body: []byte("gopher")}}
	require.NoError(t, fn(context.Background(), p))
	assert.Equal(t, testGZIPBody, p.Body)
}

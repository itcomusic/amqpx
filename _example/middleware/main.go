package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"

	"github.com/itcomusic/amqpx"
	"github.com/itcomusic/amqpx/amqpxotel"
)

func main() {
	conn, _ := amqpx.Connect(
		amqpx.UseConsumeHook(amqpxotel.Consumer(otel.Tracer(""), "amqp")),
		amqpx.UsePublishHook(amqpxotel.Publisher(otel.Tracer(""))))
	defer conn.Close()

	_ = amqpx.NewPublisher[[]byte](conn, amqpx.Direct, amqpx.SetPublishHook(func(next amqpx.PublisherFunc) amqpx.PublisherFunc {
		return func(ctx context.Context, m *amqpx.PublishRequest) error {
			fmt.Printf("message: %s\n", m.Body)
			return next(ctx, m)
		}
	}))
}

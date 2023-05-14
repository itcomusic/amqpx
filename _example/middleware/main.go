package main

import (
	"context"
	"fmt"

	"github.com/itcomusic/amqpx"
	"github.com/itcomusic/amqpx/amqpxotel"
)

func main() {
	conn, _ := amqpx.Connect(amqpx.UseInterceptor(amqpxotel.NewInterceptor()))
	defer conn.Close()

	_ = amqpx.NewPublisher[[]byte](conn, amqpx.ExchangeDirect, amqpx.SetPublishInterceptor(func(next amqpx.PublishFunc) amqpx.PublishFunc {
		return func(ctx context.Context, m *amqpx.PublishingRequest) error {
			fmt.Printf("message: %s\n", m.Body)
			return next(ctx, m)
		}
	}))
}

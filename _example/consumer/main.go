package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"

	"github.com/itcomusic/amqpx"
	"github.com/itcomusic/amqpx/amqpxjson"
	"github.com/itcomusic/amqpx/amqpxotel"
	"github.com/itcomusic/amqpx/amqpxproto"
	"github.com/itcomusic/amqpx/amqpxprotojson"
)

func main() {
	conn, _ := amqpx.Connect(
		amqpx.UseUnmarshaler(
			amqpxproto.NewUnmarshaler(),
			amqpxprotojson.NewUnmarshaler()), // global unmarshalers
		amqpx.UseConsumeHook(amqpxotel.Consumer(otel.Tracer(""), "amqp")))
	defer conn.Close()

	// []byte
	{
		_ = conn.NewConsumer("foo", amqpx.ConsumerFunc(func(d *amqpx.Delivery) amqpx.Action {
			fmt.Printf("received message: %s\n", string(d.Body))
			return amqpx.Ack
		}))
	}

	// message
	{
		type Gopher struct {
			Name string
		}

		resetFn := func(v *Gopher) { v.Name = "" } // using sync.Pool
		_ = conn.NewConsumer("bar", amqpx.ConsumerMessage(func(ctx context.Context, m *Gopher) amqpx.Action {
			fmt.Printf("user-id: %s, received message: %s\n", amqpx.FromContext(ctx).UserId, m.Name)
			return amqpx.Ack
		}, amqpx.SetPool(resetFn)), amqpx.SetUnmarshaler(amqpxjson.Unmarshaler), amqpx.SetAutoAckMode()) // individual single unmarshaler
	}
}

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
		amqpx.UseUnmarshaler( // global unmarshalers
			amqpxproto.NewUnmarshaler(),
			amqpxprotojson.NewUnmarshaler()),
		amqpx.UseConsumeHook(amqpxotel.Consumer(otel.Tracer(""), "amqp")))
	defer conn.Close()

	// []byte
	{
		_ = conn.NewConsumer("foo", amqpx.D(func(ctx context.Context, d *amqpx.Delivery[[]byte]) amqpx.Action {
			fmt.Printf("received message: %s\n", string(*d.Msg))
			return amqpx.Ack
		}))
	}

	// message
	{
		type Gopher struct {
			Name string
		}

		resetFn := func(v *Gopher) { v.Name = "" } // using sync.Pool
		_ = conn.NewConsumer("bar", amqpx.D(func(ctx context.Context, d *amqpx.Delivery[Gopher]) amqpx.Action {
			fmt.Printf("user-id: %s, received message: %s\n", d.Req.UserID, d.Msg.Name)
			return amqpx.Ack
		}, amqpx.SetPool(resetFn)), amqpx.SetUnmarshaler(amqpxjson.Unmarshaler), amqpx.SetAutoAckMode()) // individual single unmarshaler
	}
}

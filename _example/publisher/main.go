package main

import (
	"go.opentelemetry.io/otel"

	"github.com/itcomusic/amqpx"
	"github.com/itcomusic/amqpx/amqpxjson"
	"github.com/itcomusic/amqpx/amqpxotel"
	"github.com/itcomusic/amqpx/amqpxproto"
)

func main() {
	conn, _ := amqpx.Connect(
		amqpx.UseMarshaler(amqpxproto.NewMarshaler()), // global marshaler
		amqpx.UsePublishHook(amqpxotel.Publisher(otel.Tracer(""))))
	defer conn.Close()

	// []byte
	{
		pub := amqpx.NewPublisher[[]byte](conn, amqpx.ExchangeDirect, amqpx.UseRoutingKey("routing_key"))
		_ = pub.Publish(amqpx.NewPublishing([]byte("hello")).PersistentMode())
	}

	// message
	{
		type Gopher struct {
			Name string
		}
		pub := amqpx.NewPublisher[Gopher](conn, amqpx.ExchangeDirect, amqpx.SetMarshaler(amqpxjson.Marshaler)) // individual single marshaler
		_ = pub.Publish(amqpx.NewPublishing(Gopher{Name: "Rob"}), amqpx.SetRoutingKey("routing_key"))
	}
}

# RabbitMQ Go Client

[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![coverage-img]][coverage-url]

This is a Go AMQP 0.9.1 client wraps [amqp091-go](https://github.com/rabbitmq/amqp091-go)
* Support of the encoding messages
    * defaults encoding (json, protobuf, protojson)
    * support of custom marshal/unmarshal functions
* Middleware for easy integration

## Installation
Go version 1.18+
```bash
go get github.com/itcomusic/amqpx
```

## Usage
```go
package main

import (
	"fmt"

	"github.com/itcomusic/amqpx"
)

func main() {
    conn, _ := amqpx.Connect()
    defer conn.Close()

    // simple publisher
    pub := amqpx.NewPublisher[[]byte](conn, amqpx.Direct, amqpx.UseRoutingKey("routing_key"))
    _ = pub.Publish(pub.NewPublishing([]byte("hello")).PersistentMode())
	
    // simple consumer 
    _ = conn.NewConsumer("foo", amqpx.D(func(d *amqpx.Delivery) amqpx.Action {
        fmt.Printf("received message: %s\n", string(d.Body))
        return amqpx.Ack
    }))
}
```

### Publisher & consumer struct
Pretty using struct and avoiding boilerplate marhsal/unmarshal. It is strict compared content-type of the message and invalid body is rejected.
```go
    type Gopher struct {
        Name string
    }
    pub := amqpx.NewPublisher[Gopher](conn, amqpx.Direct, amqpx.SetMarshaler(amqpxjson.Marshaler))
    _ = pub.Publish(pub.NewPublishing(Gopher{Name: "Rob"}), amqpx.SetRoutingKey("routing_key"))

    resetFn := func(v *Gopher) { v.Name = "" } // option using sync.Pool
    _ = conn.NewConsumer("bar", amqpx.T(func(ctx context.Context, m *Gopher) amqpx.Action {
        fmt.Printf("user-id: %s, received message: %s\n", amqpx.FromContext(ctx).UserId, m.Name)
        return amqpx.Ack
    }, amqpx.SetPool(resetFn)), amqpx.SetUnmarshaler(amqpxjson.Unmarshaler), amqpx.SetAutoAckMode())
```

### Consumer rate limiting
The Prefetch count informs the server will deliver that many messages to consumers before acknowledgments are received. 
The Concurrency option limits numbers of goroutines of consumer, depends on prefetch count and auto-ack mode.
```go
    // prefetch count
    _ = conn.NewConsumer("foo", amqpx.D(func(d *amqpx.Delivery) amqpx.Action {
        fmt.Printf("received message: %s\n", string(d.Body))
        return amqpx.Ack
    }), amqpx.SetPrefetchCount(8))

    // limit goroutines
	_ = conn.NewConsumer("foo", amqpx.D(func(d *amqpx.Delivery) amqpx.Action {
        fmt.Printf("received message: %s\n", string(d.Body))
        return amqpx.Ack
    }), amqpx.SetAutoAckMode(), amqpx.SetConcurrency(32))
```

### Declare queue
The declare queue, exchange and binding queue.
```go
    _ = conn.NewConsumer("foo", amqpx.D(func(d *amqpx.Delivery) amqpx.Action {
        fmt.Printf("received message: %s\n", string(d.Body))
        return amqpx.Ack
    }), amqpx.DeclareQueue(amqpx.QueueDeclare{AutoDelete: true}),
        amqpx.DeclareExchange(amqpx.ExchangeDeclare{Name: "exchange_name", Type: amqpx.Direct}),
        amqpx.BindQueue(amqpx.QueueBind{Exchange: "exchange_name", RoutingKey: []string{"routing_key"}}))
```

### Middleware
Predefined support opentelemetry using hooks.
```go
    import (
        "github.com/itcomusic/amqpx"
        "github.com/itcomusic/amqpx/amqpxotel"
    )

    // global
    conn, _ := amqpx.Connect(
        amqpx.UseConsumeHook(amqpxotel.Consumer(otel.Tracer(""), "amqp")),
        amqpx.UsePublishHook(amqpxotel.Publisher(otel.Tracer(""))))
    defer conn.Close()

    // special hook
    _ = amqpx.NewPublisher[[]byte](conn, amqpx.Direct, amqpx.SetPublishHook(func(next amqpx.PublisherFunc) amqpx.PublisherFunc {
        return func(m *amqpx.Publishing) error {
            fmt.Printf("message: %s\n", m.Body)
            return next(m)
        }
    }))
```
## License
[MIT License](LICENSE)

[build-img]: https://github.com/itcomusic/amqpx/workflows/build/badge.svg
[build-url]: https://github.com/itcomusic/amqpx/actions
[pkg-img]: https://pkg.go.dev/badge/github.com/itcomusic/amqpx.svg
[pkg-url]: https://pkg.go.dev/github.com/itcomusic/amqpx
[coverage-img]: https://codecov.io/gh/itcomusic/amqpx/branch/main/graph/badge.svg
[coverage-url]: https://codecov.io/gh/itcomusic/amqpx
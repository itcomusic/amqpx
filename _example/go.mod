module github.com/itcomusic/amqpx/_example

go 1.20

replace github.com/itcomusic/amqpx => ../.

replace github.com/itcomusic/amqpx/amqpxotel => ../amqpxotel

replace github.com/itcomusic/amqpx/amqpxprotojson => ../amqpxprotojson

replace github.com/itcomusic/amqpx/amqpxproto => ../amqpxproto

require (
	github.com/itcomusic/amqpx v0.2.0
	github.com/itcomusic/amqpx/amqpxotel v0.0.0-00010101000000-000000000000
	github.com/itcomusic/amqpx/amqpxproto v0.0.0-00010101000000-000000000000
	github.com/itcomusic/amqpx/amqpxprotojson v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/rabbitmq/amqp091-go v1.9.0 // indirect
	go.opentelemetry.io/otel v1.17.0 // indirect
	go.opentelemetry.io/otel/metric v1.17.0 // indirect
	go.opentelemetry.io/otel/trace v1.17.0 // indirect
	golang.org/x/sync v0.4.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

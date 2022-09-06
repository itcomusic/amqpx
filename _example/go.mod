module github.com/itcomusic/amqpx/_example

go 1.19

replace github.com/itcomusic/amqpx => ../.

replace github.com/itcomusic/amqpx/amqpxotel => ../amqpxotel

replace github.com/itcomusic/amqpx/amqpxprotojson => ../amqpxprotojson

replace github.com/itcomusic/amqpx/amqpxproto => ../amqpxproto

require (
	github.com/itcomusic/amqpx v0.5.3
	github.com/itcomusic/amqpx/amqpxotel v0.0.0-00010101000000-000000000000
	github.com/itcomusic/amqpx/amqpxproto v0.0.0-00010101000000-000000000000
	github.com/itcomusic/amqpx/amqpxprotojson v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.7.0
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/rabbitmq/amqp091-go v1.3.4 // indirect
	go.opentelemetry.io/otel/trace v1.7.0 // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

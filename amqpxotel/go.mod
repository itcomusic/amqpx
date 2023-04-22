module github.com/itcomusic/amqpx/amqpxotel

go 1.19

replace github.com/itcomusic/amqpx => ../.

require (
	github.com/itcomusic/amqpx v0.2.0
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/trace v1.14.0
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/rabbitmq/amqp091-go v1.8.0 // indirect
	golang.org/x/sync v0.0.0-20220819030929-7fc1605a5dde // indirect
)

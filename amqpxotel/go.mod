module github.com/itcomusic/amqpx/amqpxotel

go 1.18

replace github.com/itcomusic/amqpx => ../.

require (
	github.com/itcomusic/amqpx v0.0.0-20220714081720-edb651d84bcd
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/rabbitmq/amqp091-go v1.3.4 // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
)

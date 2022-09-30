module github.com/itcomusic/amqpx/amqpxzap

go 1.19

replace github.com/itcomusic/amqpx => ../.

require (
	github.com/itcomusic/amqpx v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.19.1
)

require (
	github.com/rabbitmq/amqp091-go v1.5.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sync v0.0.0-20220819030929-7fc1605a5dde // indirect
)

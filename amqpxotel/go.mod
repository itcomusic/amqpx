module github.com/itcomusic/amqpx/amqpxotel

go 1.20

replace github.com/itcomusic/amqpx => ../.

require (
	github.com/itcomusic/amqpx v0.2.0
	github.com/rabbitmq/amqp091-go v1.8.1
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/otel v1.17.0
	go.opentelemetry.io/otel/sdk v1.17.0
	go.opentelemetry.io/otel/trace v1.17.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/otel/metric v1.17.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

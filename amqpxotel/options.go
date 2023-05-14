package amqpxotel

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/itcomusic/amqpx"
)

type config struct {
	consumeFilter func(context.Context, *amqpx.DeliveryRequest) bool
	operationName func(context.Context, *amqpx.DeliveryRequest) string
	tracer        trace.Tracer
	propagator    propagation.TextMapPropagator
}

// An Option configures the OpenTelemetry instrumentation.
type Option func(*config)

// WithPropagator configures the instrumentation to use the supplied propagator
// when extracting and injecting trace context. By default, the instrumentation
// uses otel.GetTextMapPropagator().
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(c *config) {
		if propagator != nil {
			c.propagator = propagator
		}
	}
}

// WithTracerProvider configures the instrumentation to use the supplied
// provider when creating a tracer. By default, the instrumentation
// uses otel.GetTracerProvider().
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(c *config) {
		if provider != nil {
			c.tracer = provider.Tracer(
				instrumentationName,
				trace.WithInstrumentationVersion(semanticVersion),
			)
		}
	}
}

// WithConsumeFilter configures the instrumentation to emit traces when the filter function returns true.
func WithConsumeFilter(filter func(context.Context, *amqpx.DeliveryRequest) bool) Option {
	return func(c *config) {
		if filter != nil {
			c.consumeFilter = filter
		}
	}
}

// WithOperationName configures the traces with the operation name.
func WithOperationName(operationName func(context.Context, *amqpx.DeliveryRequest) string) Option {
	return func(c *config) {
		if operationName != nil {
			c.operationName = operationName
		}
	}
}

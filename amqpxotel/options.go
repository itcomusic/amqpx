package amqpxotel

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/itcomusic/amqpx"
)

type (
	consumeFilter   func(context.Context, *amqpx.DeliveryRequest) bool
	publishFilter   func(context.Context, *amqpx.PublishingRequest) bool
	consumeSpanName func(context.Context, *amqpx.DeliveryRequest) string
	publishSpanName func(context.Context, *amqpx.PublishingRequest) string
	ackStatus       func(amqpx.Action) (code codes.Code, description string)
)

type config struct {
	tracer          trace.Tracer
	propagator      propagation.TextMapPropagator
	consumeFilter   consumeFilter
	publishFilter   publishFilter
	consumeSpanName consumeSpanName
	publishSpanName publishSpanName
	ackStatus       ackStatus
	notTrustRemote  bool
}

// An Option configuring the OpenTelemetry instrumentation.
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

// WithConsumeFilter configures the Interceptor to emit traces when the filter function returns true.
func WithConsumeFilter(filter func(context.Context, *amqpx.DeliveryRequest) bool) Option {
	return func(c *config) {
		if filter != nil {
			c.consumeFilter = filter
		}
	}
}

// WithPublishFilter configures the Interceptor to emit traces when the filter function returns true.
func WithPublishFilter(filter func(context.Context, *amqpx.PublishingRequest) bool) Option {
	return func(c *config) {
		if filter != nil {
			c.publishFilter = filter
		}
	}
}

// WithConsumeSpanName configures the Interceptor to span name.
func WithConsumeSpanName(operationName func(context.Context, *amqpx.DeliveryRequest) string) Option {
	return func(c *config) {
		if operationName != nil {
			c.consumeSpanName = operationName
		}
	}
}

// WithPublishSpanName configures the Interceptor span name.
func WithPublishSpanName(operationName func(context.Context, *amqpx.PublishingRequest) string) Option {
	return func(c *config) {
		if operationName != nil {
			c.publishSpanName = operationName
		}
	}
}

// WithAckStatus configures the Interceptor to handle the acknowledgment status.
// By default, amqpx.Ack sets codes.Ok and other codes.Unset.
func WithAckStatus(ackStatus func(amqpx.Action) (code codes.Code, description string)) Option {
	return func(c *config) {
		if ackStatus != nil {
			c.ackStatus = ackStatus
		}
	}
}

// WithNotTrustRemote sets the Interceptor to not trust remote spans.
// By default, all incoming spans are trusted and will be a child span.
// When WithNotTrustRemote is used, all incoming spans are untrusted and will be linked
// with a [trace.Link] and will not be a child span.
func WithNotTrustRemote() Option {
	return func(c *config) {
		c.notTrustRemote = true
	}
}

func defaultAckStatus(action amqpx.Action) (code codes.Code, description string) {
	if amqpx.Ack == action {
		return codes.Ok, ""
	}
	return codes.Unset, ""
}

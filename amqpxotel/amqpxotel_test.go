package amqpxotel

import (
	"context"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	"github.com/itcomusic/amqpx"
)

func TestInterceptor_WrapConsume(t *testing.T) {
	t.Parallel()

	spanRecorder := tracetest.NewSpanRecorder()
	traceProvider := trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))

	fn := NewInterceptor(WithTracerProvider(traceProvider)).WrapConsume(func(ctx context.Context, req *amqpx.DeliveryRequest) amqpx.Action {
		return amqpx.Ack
	})

	d := (&amqpx.DeliveryRequest{}).NewFrom(&amqp091.Delivery{Exchange: "direct", RoutingKey: "key"})
	status := fn(context.Background(), d)
	require.Equal(t, amqpx.Ack, status)
	assertSpans(t, []wantSpans{{
		spanName: defaultName("direct", "key", messagingOperationProcess),
		attrs: []attribute.KeyValue{
			semconv.MessagingSystem(messagingSystem),
			semconv.MessagingOperationProcess,
			semconv.MessagingMessagePayloadSizeBytes(0),
			semconv.MessagingDestinationName("direct"),
			semconv.MessagingRabbitmqDestinationRoutingKey("key"),
		},
	}}, spanRecorder.Ended())
}

func TestInterceptor_WrapPublish(t *testing.T) {
	t.Parallel()

	spanRecorder := tracetest.NewSpanRecorder()
	traceProvider := trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))

	fn := NewInterceptor(WithTracerProvider(traceProvider)).WrapPublish(func(ctx context.Context, req *amqpx.PublishingRequest) error {
		return nil
	})

	req := (&amqpx.PublishingRequest{Publishing: amqp091.Publishing{Headers: make(amqp091.Table)}}).NewFrom("direct", "key")
	err := fn(context.Background(), req)
	require.NoError(t, err)
	assertSpans(t, []wantSpans{{
		spanName: defaultName("direct", "key", messagingOperationPublish),
		attrs: []attribute.KeyValue{
			semconv.MessagingSystem(messagingSystem),
			semconv.MessagingOperationPublish,
			semconv.MessagingMessagePayloadSizeBytes(0),
			semconv.MessagingDestinationName("direct"),
			semconv.MessagingRabbitmqDestinationRoutingKey("key"),
		},
	}}, spanRecorder.Ended())
}

type wantSpans struct {
	spanName string
	events   []trace.Event
	attrs    []attribute.KeyValue
}

func assertSpans(t *testing.T, want []wantSpans, got []trace.ReadOnlySpan) {
	t.Helper()
	assert.Equal(t, len(want), len(got), "span count")

	for i, span := range got {
		wantEvents := want[i].events
		wantAttributes := want[i].attrs

		require.Equal(t, span.EndTime().IsZero(), false, "span end time")
		assert.Equal(t, want[i].spanName, span.Name(), "span name")
		assert.ElementsMatch(t, wantEvents, span.Events(), "events")
		assert.ElementsMatch(t, wantAttributes, span.Attributes(), "attributes")
	}
}

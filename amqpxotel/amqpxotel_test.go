package amqpxotel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/itcomusic/amqpx"
)

func TestInterceptor(t *testing.T) {
	t.Parallel()

	spanRecorder := tracetest.NewSpanRecorder()
	traceProvider := trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))

	fn := NewInterceptor(WithTracerProvider(traceProvider)).WrapConsume(func(ctx context.Context, req *amqpx.DeliveryRequest) amqpx.Action {
		return amqpx.Ack
	})
	status := fn(context.Background(), &amqpx.DeliveryRequest{Exchange: "direct", RoutingKey: "key"})
	assert.Equal(t, amqpx.Ack, status)
	assertSpans(t, []wantSpans{{
		spanName: "direct.key",
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
		assert.Equal(t, wantEvents, span.Events(), "events")
		assert.Equal(t, wantAttributes, span.Attributes(), "attributes")
	}
}

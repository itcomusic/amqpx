// Package amqpxotel provides supporting opentracing.
package amqpxotel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/itcomusic/amqpx"
)

// Consumer returns consume hook that wraps the next.
func Consumer(tracer trace.Tracer, operationName string) amqpx.ConsumeHook {
	return func(next amqpx.ConsumeFunc) amqpx.ConsumeFunc {
		return func(ctx context.Context, req *amqpx.DeliveryRequest) amqpx.Action {
			spanContext := otel.GetTextMapPropagator().Extract(ctx, table(req.Headers))
			ctx, span := tracer.Start(ctx, operationName, trace.WithSpanKind(trace.SpanKindConsumer), trace.WithLinks(trace.LinkFromContext(spanContext)))
			defer span.End()
			return next(ctx, req)
		}
	}
}

// Publisher returns publish hook that wraps the next.
func Publisher(tracer trace.Tracer) amqpx.PublishHook {
	return func(next amqpx.PublisherFunc) amqpx.PublisherFunc {
		return func(ctx context.Context, m *amqpx.PublishRequest) error {
			otel.GetTextMapPropagator().Inject(ctx, table(m.Headers))
			return next(ctx, m)
		}
	}
}

type table map[string]any

var _ propagation.TextMapCarrier = (*table)(nil)

func (t table) Get(key string) string {
	v, ok := t[key].(string)
	if !ok {
		return ""
	}
	return v
}

func (t table) Keys() []string {
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	return keys
}

func (t table) Set(key, val string) {
	t[key] = val
}

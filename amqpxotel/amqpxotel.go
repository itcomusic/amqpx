// Package amqpxotel provides hooks for opentracing.
package amqpxotel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/itcomusic/amqpx"
)

// Consumer returns consume hook that wraps the next.
func Consumer(tracer trace.Tracer, operationName string) amqpx.ConsumeHook {
	return func(next amqpx.Consumer) amqpx.Consumer {
		return amqpx.ConsumerFunc(func(d *amqpx.Delivery) amqpx.Action {
			spanContext := otel.GetTextMapPropagator().Extract(d.Context(), table(d.Headers))
			ctx, span := tracer.Start(d.Context(), operationName, trace.WithSpanKind(trace.SpanKindConsumer), trace.WithLinks(trace.LinkFromContext(spanContext)))
			defer span.End()
			d.WithContext(ctx)
			return next.Serve(d)
		})
	}
}

// Publisher returns publish hook that wraps the next.
func Publisher(tracer trace.Tracer) amqpx.PublishHook {
	return func(next amqpx.PublisherFunc) amqpx.PublisherFunc {
		return func(m *amqpx.Publishing) error {
			otel.GetTextMapPropagator().Inject(m.Context(), table(m.Headers))
			return next(m)
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

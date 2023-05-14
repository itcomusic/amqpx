package amqpxotel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/itcomusic/amqpx"
)

type Interceptor struct {
	config config
}

var _ amqpx.Interceptor = (*Interceptor)(nil)

// NewInterceptor returns a new otel interceptor.
func NewInterceptor(opts ...Option) *Interceptor {
	cfg := config{
		tracer: otel.GetTracerProvider().Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(semanticVersion),
		),
		propagator: otel.GetTextMapPropagator(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Interceptor{
		config: cfg,
	}
}

func (i Interceptor) WrapConsume(next amqpx.ConsumeFunc) amqpx.ConsumeFunc {
	return func(ctx context.Context, req *amqpx.DeliveryRequest) amqpx.Action {
		if i.config.consumeFilter != nil {
			if i.config.consumeFilter(ctx, req) {
				return next(ctx, req)
			}
		}

		operationName := req.Exchange + "." + req.RoutingKey
		if i.config.operationName != nil {
			operationName = i.config.operationName(ctx, req)
		}

		spanContext := i.config.propagator.Extract(ctx, table(req.Headers))
		ctx, span := i.config.tracer.Start(ctx, operationName, trace.WithSpanKind(trace.SpanKindConsumer), trace.WithLinks(trace.LinkFromContext(spanContext)))
		defer span.End()
		return next(ctx, req)
	}
}

func (i Interceptor) WrapPublish(next amqpx.PublishFunc) amqpx.PublishFunc {
	return func(ctx context.Context, m *amqpx.PublishingRequest) error {
		otel.GetTextMapPropagator().Inject(ctx, table(m.Headers))
		return next(ctx, m)
	}
}

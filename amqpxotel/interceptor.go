package amqpxotel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/itcomusic/amqpx"
)

const (
	messagingSystem           = "rabbitmq"
	messagingOperationProcess = "process"
	messagingOperationPublish = "publish"
	messagingClientIDKey      = attribute.Key("messaging.rabbitmq.client_id")
	messagingAppIDKey         = attribute.Key("messaging.rabbitmq.app_id")
	messagingReplyToKey       = attribute.Key("messaging.rabbitmq.reply_to")
	messagingRedeliveredKey   = attribute.Key("messaging.rabbitmq.redelivered")
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
		ackStatus:  defaultAckStatus,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Interceptor{
		config: cfg,
	}
}

func (i *Interceptor) WrapConsume(next amqpx.ConsumeFunc) amqpx.ConsumeFunc {
	return func(ctx context.Context, req *amqpx.DeliveryRequest) amqpx.Action {
		if i.config.consumeFilter != nil {
			if i.config.consumeFilter(ctx, req) {
				return next(ctx, req)
			}
		}

		// attrs
		attr := []attribute.KeyValue{
			semconv.MessagingSystem(messagingSystem),
			semconv.MessagingOperationProcess,
			semconv.MessagingMessagePayloadSizeBytes(len(req.Body())),
		}
		if v := req.Exchange(); v != "" {
			attr = append(attr, semconv.MessagingDestinationName(v))
		}
		if v := req.RoutingKey(); v != "" {
			attr = append(attr, semconv.MessagingRabbitmqDestinationRoutingKey(v))
		}
		if v := req.AppID(); v != "" {
			attr = append(attr, messagingAppIDKey.String(v))
		}
		if v := req.MessageID(); v != "" {
			attr = append(attr, semconv.MessagingMessageID(v))
		}
		if v := req.Redelivered(); v {
			attr = append(attr, messagingRedeliveredKey.Bool(true))
		}
		if v := req.ConsumerTag(); v != "" {
			attr = append(attr, semconv.MessagingConsumerID(v))
		}
		if v := req.UserID(); v != "" {
			attr = append(attr, messagingClientIDKey.String(v))
		}
		if v := req.CorrelationID(); v != "" {
			attr = append(attr, semconv.MessagingMessageConversationID(v))
		}
		if v := req.ReplyTo(); v != "" {
			attr = append(attr, messagingReplyToKey.String(v))
		}

		operationName := defaultName(req.Exchange(), req.RoutingKey(), messagingOperationProcess)
		if i.config.consumeSpanName != nil {
			operationName = i.config.consumeSpanName(ctx, req)
		}

		traceOpts := []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(attr...),
		}

		// if a span already exists in ctx, then there must have already been another interceptor
		// that set it, so don't extract from carrier.
		if !trace.SpanContextFromContext(ctx).IsValid() {
			ctx = i.config.propagator.Extract(ctx, table(req.Headers()))
			if i.config.notTrustRemote {
				traceOpts = append(traceOpts,
					trace.WithNewRoot(),
					trace.WithLinks(trace.LinkFromContext(ctx)),
				)
			}
		}

		ctx, span := i.config.tracer.Start(ctx, operationName, traceOpts...)
		defer span.End()

		status := next(ctx, req)
		code, desc := i.config.ackStatus(status)
		span.SetStatus(code, desc)
		return status
	}
}

func (i *Interceptor) WrapPublish(next amqpx.PublishFunc) amqpx.PublishFunc {
	return func(ctx context.Context, req *amqpx.PublishingRequest) error {
		if i.config.publishFilter != nil {
			if i.config.publishFilter(ctx, req) {
				return next(ctx, req)
			}
		}

		// attrs
		attr := []attribute.KeyValue{
			semconv.MessagingSystem(messagingSystem),
			semconv.MessagingOperationPublish,
			semconv.MessagingMessagePayloadSizeBytes(len(req.Body)),
		}
		if v := req.Exchange(); v != "" {
			attr = append(attr, semconv.MessagingDestinationName(v))
		}
		if v := req.RoutingKey(); v != "" {
			attr = append(attr, semconv.MessagingRabbitmqDestinationRoutingKey(v))
		}
		if req.AppId != "" {
			attr = append(attr, messagingAppIDKey.String(req.AppId))
		}
		if req.MessageId != "" {
			attr = append(attr, semconv.MessagingMessageID(req.MessageId))
		}
		if req.UserId != "" {
			attr = append(attr, messagingClientIDKey.String(req.UserId))
		}
		if req.CorrelationId != "" {
			attr = append(attr, semconv.MessagingMessageConversationID(req.CorrelationId))
		}
		if req.ReplyTo != "" {
			attr = append(attr, messagingReplyToKey.String(req.ReplyTo))
		}

		operationName := defaultName(req.Exchange(), req.RoutingKey(), messagingOperationPublish)
		if i.config.publishSpanName != nil {
			operationName = i.config.publishSpanName(ctx, req)
		}

		// inject span context into carrier
		ctx, span := i.config.tracer.Start(ctx, operationName,
			trace.WithSpanKind(trace.SpanKindProducer),
			trace.WithAttributes(attr...))
		defer span.End()
		i.config.propagator.Inject(ctx, table(req.Headers))

		err := next(ctx, req)
		if err != nil {
			span.RecordError(err)
		}
		return err
	}
}

func defaultName(exchange, key, operationType string) string {
	if exchange != "" {
		return exchange + "." + key + " " + operationType
	}
	return key + " " + operationType
}

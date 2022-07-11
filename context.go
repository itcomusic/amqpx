package amqpx

import "context"

type contextDelivery struct{}

// FromContext returns Delivery from context.
func FromContext(ctx context.Context) *Delivery {
	res, ok := ctx.Value(contextDelivery{}).(*Delivery)
	if !ok {
		return nil
	}
	return res
}

// toContext returns context with Delivery.
func toContext(d *Delivery) context.Context {
	return context.WithValue(d.ctx, contextDelivery{}, d)
}

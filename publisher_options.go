package amqpx

import "context"

// PublisherOption is used to configure a publisher.
type PublisherOption func(*publisherOptions)

type publisherOptions struct {
	confirm   bool
	publish   publishOptions
	marshaler Marshaler
	hook      []PublishHook
}

func (p *publisherOptions) validate() error {
	if p.marshaler == nil {
		return errMarshalerNotFound
	}
	return nil
}

// SetConfirmMode sets confirm mode.
//
// confirm sets channel into confirm mode so that the client can ensure
// all publishers have successfully been received by the server.
func SetConfirmMode() PublisherOption {
	return func(o *publisherOptions) {
		o.confirm = true
	}
}

// SetPublishHook sets publish hook.
func SetPublishHook(h ...PublishHook) PublisherOption {
	return func(o *publisherOptions) {
		o.hook = append(o.hook, h...)
	}
}

// SetMarshaler sets marshaler.
func SetMarshaler(m Marshaler) PublisherOption {
	return func(o *publisherOptions) {
		if m != nil {
			o.marshaler = m
		}
	}
}

// UseRoutingKey sets routing key.
func UseRoutingKey(s string) PublisherOption {
	return func(o *publisherOptions) {
		if s != "" {
			o.publish.key = s
		}
	}
}

// UseMandatory sets mandatory.
// The default is false.
//
// Message can be undeliverable when the mandatory flag is true and no queue is
// bound that matches the routing key.
func UseMandatory(b bool) PublisherOption {
	return func(o *publisherOptions) {
		o.publish.mandatory = b
	}
}

// UseImmediate sets immediate.
// The default is false.
//
// Message can be undeliverable when the immediate flag is true and no
// consumer on the matched queue is ready to accept the delivery.
func UseImmediate(b bool) PublisherOption {
	return func(o *publisherOptions) {
		o.publish.immediate = b
	}
}

// PublishOption is used to configure the publishing message.
type PublishOption func(*publishOptions)

type publishOptions struct {
	key       string
	mandatory bool
	immediate bool
	ctx       context.Context
}

func (p publishOptions) validate() error {
	if p.key == "" {
		return errRoutingKeyEmpty
	}
	return nil
}

// SetRoutingKey sets routing key.
func SetRoutingKey(s string) PublishOption {
	return func(o *publishOptions) {
		if s != "" {
			o.key = s
		}
	}
}

// SetMandatory sets mandatory.
// The default is false.
//
// Message can be undeliverable when the mandatory flag is true and no queue is
// bound that matches the routing key.
func SetMandatory(b bool) PublishOption {
	return func(o *publishOptions) {
		o.mandatory = b
	}
}

// SetImmediate sets immediate.
// The default is false.
//
// Message can be undeliverable when the immediate flag is true and no
// consumer on the matched queue is ready to accept the delivery.
func SetImmediate(b bool) PublishOption {
	return func(o *publishOptions) {
		o.immediate = b
	}
}

// SetContext sets publish context.
func SetContext(ctx context.Context) PublishOption {
	return func(o *publishOptions) {
		if ctx != nil {
			o.ctx = ctx
		}
	}
}

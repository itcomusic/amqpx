package amqpx

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
		return ErrMarshalerNotFound
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

// WithPublishOptions sets publish options.
func WithPublishOptions(opts ...PublishOption) PublisherOption {
	return func(o *publisherOptions) {
		for _, v := range opts {
			v(&o.publish)
		}
	}
}

// PublishOption is used to configure the publish message.
type PublishOption func(*publishOptions)

type publishOptions struct {
	key       string
	mandatory bool
	immediate bool
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

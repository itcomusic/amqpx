package amqpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublisherOptions(t *testing.T) {
	t.Parallel()

	got := &publisherOptions{}
	for _, o := range []PublisherOption{
		SetConfirmMode(),
		SetMarshaler(defaultBytesMarshaler),
		WithPublishOptions(
			SetRoutingKey("key"),
			SetMandatory(true),
			SetImmediate(true)),
	} {
		o(got)
	}

	want := &publisherOptions{
		confirm: true,
		publish: publishOptions{
			key:       "key",
			mandatory: true,
			immediate: true,
		},
		marshaler: defaultBytesMarshaler,
	}
	assert.Equal(t, want, got)
}

func TestPublisherOptions_Validate(t *testing.T) {
	t.Parallel()

	got := (&publisherOptions{}).validate()
	assert.ErrorIs(t, got, ErrMarshalerNotFound)
}

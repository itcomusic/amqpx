package amqpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublisherOption(t *testing.T) {
	t.Parallel()

	got := &publisherOptions{}
	for _, o := range []PublisherOption{
		SetConfirmMode(),
		SetMarshaler(defaultBytesMarshaler),
		UseRoutingKey("key"),
		UseMandatory(true),
		UseImmediate(true),
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

func TestPublisherOption_Validate(t *testing.T) {
	t.Parallel()

	got := (&publisherOptions{}).validate()
	assert.ErrorIs(t, got, ErrMarshalerNotFound)
}

package amqpx

import (
	"context"
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

	t.Run("marshaler", func(t *testing.T) {
		t.Parallel()

		got := (&publisherOptions{}).validate()
		assert.ErrorIs(t, got, errMarshalerNotFound)
	})

	t.Run("routing-key", func(t *testing.T) {
		t.Parallel()

		got := (&publishOptions{}).validate()
		assert.ErrorIs(t, got, errRoutingKeyEmpty)
	})
}

func TestPublishOptions(t *testing.T) {
	t.Parallel()

	got := &publishOptions{}
	for _, o := range []PublishOption{
		SetRoutingKey("key"),
		SetMandatory(true),
		SetImmediate(true),
		SetContext(context.Background()),
	} {
		o(got)
	}

	want := &publishOptions{
		key:       "key",
		mandatory: true,
		immediate: true,
		ctx:       context.Background(),
	}
	assert.Equal(t, want, got)
}

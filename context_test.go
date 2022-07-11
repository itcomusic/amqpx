package amqpx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, (*Delivery)(nil), FromContext(context.Background()))
	})

	t.Run("not nil", func(t *testing.T) {
		t.Parallel()
		d := &Delivery{}
		assert.Equal(t, d, FromContext(context.WithValue(context.Background(), contextDelivery{}, d)))
	})
}

func TestToContext(t *testing.T) {
	t.Parallel()

	d := &Delivery{ctx: context.Background()}
	want := context.WithValue(context.Background(), contextDelivery{}, d)
	assert.Equal(t, want, toContext(d))
}

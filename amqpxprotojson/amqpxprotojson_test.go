package amqpxprotojson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:generate protoc --go_out=/tmp amqpxprotojson.proto
//go:generate sh -c "mv /tmp/amqpxprotojson/amqpxprotojson.pb.go amqpxprotojson_pb_test.go"

func TestMarshaler_Marshal(t *testing.T) {
	t.Parallel()

	t.Run("pointer", func(t *testing.T) {
		got, err := NewMarshaler().Marshal(&Gopher{Name: "gopher"})
		assert.NoError(t, err)
		assert.Equal(t, `{"name":"gopher"}`, string(got))
	})

	t.Run("no pointer", func(t *testing.T) {
		_, err := NewMarshaler().Marshal(Gopher{Name: "gopher"})
		assert.ErrorIs(t, err, ErrProtoType)
	})
}

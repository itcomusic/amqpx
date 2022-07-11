package amqpxproto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:generate protoc --go_out=/tmp amqpxproto.proto
//go:generate sh -c "mv /tmp/amqpxproto/amqpxproto.pb.go amqpxproto_pb_test.go"

func TestMarshaler_Marshal(t *testing.T) {
	t.Parallel()

	t.Run("pointer", func(t *testing.T) {
		_, err := NewMarshaler().Marshal(&Gopher{Name: "gopher"})
		assert.NoError(t, err)
	})

	t.Run("no pointer", func(t *testing.T) {
		_, err := NewMarshaler().Marshal(Gopher{Name: "gopher"})
		assert.ErrorIs(t, err, ErrProtoType)
	})
}

package amqpxjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshaler_Marshal(t *testing.T) {
	t.Parallel()

	type Gopher struct {
		Name string `json:"name"`
	}

	got, err := Marshaler.Marshal(Gopher{Name: "gopher"})
	require.NoError(t, err)
	assert.Equal(t, `{"name":"gopher"}`, string(got))
}

func TestMarshaler_Unmarshal(t *testing.T) {
	t.Parallel()

	type Gopher struct {
		Name string `json:"name"`
	}
	require.NoError(t, Unmarshaler.Unmarshal([]byte(`{"name":"gopher"}`), &Gopher{}))
}

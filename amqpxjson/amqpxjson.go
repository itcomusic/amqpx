package amqpxjson

import (
	"encoding/json"

	"github.com/itcomusic/amqpx"
)

var (
	_ amqpx.Marshaler   = (*marshaler)(nil)
	_ amqpx.Unmarshaler = (*unmarshaler)(nil)
)

const contentType = "application/json"

var (
	Marshaler   = (*marshaler)(nil)
	Unmarshaler = (*unmarshaler)(nil)
)

type marshaler struct{}

func (*marshaler) ContentType() string {
	return contentType
}

func (*marshaler) Marshal(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type unmarshaler struct{}

func (*unmarshaler) ContentType() string {
	return contentType
}

func (*unmarshaler) Unmarshal(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

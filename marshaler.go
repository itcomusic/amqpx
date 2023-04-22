package amqpx

import "fmt"

const bytesContentType = "application/octet-stream"

// Marshaler is an interface implemented by format.
type Marshaler interface {
	ContentType() string
	Marshal(any) ([]byte, error)
}

// Unmarshaler is an interface implemented by format.
type Unmarshaler interface {
	ContentType() string
	Unmarshal([]byte, any) error
}

var defaultBytesMarshaler = &bytesMarshaler{}

type bytesMarshaler struct{}

func (bytesMarshaler) Marshal(v any) ([]byte, error) {
	b, ok := v.(*[]byte)
	if !ok {
		return nil, fmt.Errorf("amqpx: bytes marshaler: unsupported type %T", v)
	}

	if b == nil {
		return []byte(nil), nil
	}
	return *b, nil
}

func (bytesMarshaler) ContentType() string {
	return bytesContentType
}

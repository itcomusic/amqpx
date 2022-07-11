// Package amqpxprotojson provides encoding body of the message using protojson.
package amqpxprotojson

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/itcomusic/amqpx"
)

var (
	ErrProtoType = fmt.Errorf("amqpxprotojson: expecting a value type proto.Message")

	DefaultMarshalOptions   = &protojson.MarshalOptions{}
	DefaultUnmarshalOptions = &protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

var (
	_ amqpx.Marshaler   = (*Marshaler)(nil)
	_ amqpx.Unmarshaler = (*Unmarshaler)(nil)
)

const contentType = "application/proto-json"

type Marshaler struct {
	opts *protojson.MarshalOptions
}

func NewMarshaler(opts ...protojson.MarshalOptions) *Marshaler {
	o := DefaultMarshalOptions
	for i := range opts {
		o = &opts[i]
	}

	return &Marshaler{opts: o}
}

func (m *Marshaler) ContentType() string {
	return contentType
}

func (m *Marshaler) Marshal(v any) ([]byte, error) {
	p, ok := v.(proto.Message)
	if !ok {
		return nil, ErrProtoType
	}

	b, err := m.opts.Marshal(p)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type Unmarshaler struct {
	opts *protojson.UnmarshalOptions
}

func NewUnmarshaler(opts ...protojson.UnmarshalOptions) *Unmarshaler {
	o := DefaultUnmarshalOptions
	for i := range opts {
		o = &opts[i]
	}
	return &Unmarshaler{opts: o}
}

func (u *Unmarshaler) ContentType() string {
	return contentType
}

func (u *Unmarshaler) Unmarshal(b []byte, v any) error {
	p, ok := v.(proto.Message)
	if !ok {
		return ErrProtoType
	}
	return u.opts.Unmarshal(b, p)
}

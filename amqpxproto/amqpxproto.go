// Package amqpxproto provides encoding body of the message using proto.
package amqpxproto

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/itcomusic/amqpx"
)

var (
	ErrProtoType = fmt.Errorf("amqpxproto: expecting a value type proto.Message")

	DefaultMarshalOptions   = &proto.MarshalOptions{}
	DefaultUnmarshalOptions = &proto.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

var (
	_ amqpx.Marshaler   = (*Marshaler)(nil)
	_ amqpx.Unmarshaler = (*Unmarshaler)(nil)
)

const contentType = "application/proto"

type Marshaler struct {
	opts *proto.MarshalOptions
}

func NewMarshaler(opts ...proto.MarshalOptions) *Marshaler {
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
	opts *proto.UnmarshalOptions
}

func NewUnmarshaler(opts ...proto.UnmarshalOptions) *Unmarshaler {
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

// Package amqpxotel provides supporting opentracing.
package amqpxotel

import (
	"go.opentelemetry.io/otel/propagation"
)

const (
	version             = "0.0.1"
	semanticVersion     = "semver:" + version
	instrumentationName = "github.com/itcomusic/amqpx/amqpxotel"
)

type table map[string]any

var _ propagation.TextMapCarrier = (*table)(nil)

func (t table) Get(key string) string {
	v, ok := t[key].(string)
	if !ok {
		return ""
	}
	return v
}

func (t table) Keys() []string {
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	return keys
}

func (t table) Set(key, val string) {
	t[key] = val
}

package amqpx

import (
	"crypto/tls"
	"strconv"
	"strings"

	"github.com/rabbitmq/amqp091-go"
)

type Config = amqp091.Config

type Interceptor interface {
	WrapConsume(ConsumeFunc) ConsumeFunc
	WrapPublish(PublishFunc) PublishFunc
}

// ClientOption is used to configure a client.
type ClientOption func(*clientOptions)

type clientOptions struct {
	uri    amqp091.URI
	config amqp091.Config

	marshaler   Marshaler
	unmarshaler map[string]Unmarshaler
	wrapConsume []ConsumeInterceptor
	wrapPublish []PublishInterceptor
	logger      LogFunc

	dialer dialer
	err    error
}

func newClientOptions() clientOptions {
	return clientOptions{
		uri:         defaultURI,
		config:      defaultConfig,
		unmarshaler: make(map[string]Unmarshaler),
		marshaler:   defaultBytesMarshaler,
	}
}

// ApplyURI sets amqp URI.
func ApplyURI(s string) ClientOption {
	return func(o *clientOptions) {
		u, err := amqp091.ParseURI(s)
		if err != nil {
			o.err = err
			return
		}
		o.uri = u
	}
}

// SetConfig sets amqp config.
func SetConfig(c Config) ClientOption {
	return func(o *clientOptions) {
		o.config = c
	}
}

// SetHost sets host and port.
func SetHost(h string) ClientOption {
	return func(o *clientOptions) {
		o.uri.Host = h
		hp := strings.Split(h, ":")
		if p, err := strconv.Atoi(hp[len(hp)-1]); err == nil {
			o.uri.Host = strings.Join(hp[:len(hp)-1], ":")
			o.uri.Port = p
			return
		}
	}
}

// SetVHost sets vhost.
func SetVHost(vhost string) ClientOption {
	return func(o *clientOptions) {
		o.uri.Vhost = vhost
	}
}

// SetAuth sets auth username and password.
func SetAuth(username, password string) ClientOption {
	return func(o *clientOptions) {
		o.uri.Username = username
		o.uri.Password = password
	}
}

// SetTLSConfig sets tls config.
func SetTLSConfig(t *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.config.TLSClientConfig = t
	}
}

// IsTLS sets TLS.
func IsTLS(v bool) ClientOption {
	return func(o *clientOptions) {
		if v {
			o.uri.Scheme = "amqps"
		}
	}
}

// SetConnectionName sets client connection name.
func SetConnectionName(name string) ClientOption {
	return func(o *clientOptions) {
		o.config.Properties.SetClientConnectionName(name)
	}
}

// WithLog sets log.
// The default is stdout.
func WithLog(log LogFunc) ClientOption {
	return func(o *clientOptions) {
		if o.logger != nil {
			o.logger = log
		}
	}
}

// UseInterceptor sets interceptors.
func UseInterceptor(i ...Interceptor) ClientOption {
	return func(o *clientOptions) {
		for _, v := range i {
			o.wrapConsume = append(o.wrapConsume, v.WrapConsume)
			o.wrapPublish = append(o.wrapPublish, v.WrapPublish)
		}

	}
}

// UseUnmarshaler sets unmarshaler of the consumer message.
func UseUnmarshaler(u ...Unmarshaler) ClientOption {
	return func(o *clientOptions) {
		for i, v := range u {
			if v != nil {
				o.unmarshaler[v.ContentType()] = u[i]
			}
		}
	}
}

// UseMarshaler sets marshaler of the publisher message.
func UseMarshaler(m Marshaler) ClientOption {
	return func(o *clientOptions) {
		if m != nil {
			o.marshaler = m
		}
	}
}

func setDialer(d dialer) ClientOption {
	return func(o *clientOptions) {
		o.dialer = d
	}
}

func (c *clientOptions) validate() error {
	if c.err != nil {
		return c.err
	}

	if c.logger == nil {
		c.logger = defaultLogger
	}

	if c.dialer == nil {
		c.dialer = &defaultDialer{
			URI:    c.uri.String(),
			Config: c.config,
			log:    c.logger,
		}
	}
	return nil
}

package amqpx

import (
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientOption(t *testing.T) {
	t.Parallel()

	got := newClientOptions()
	for _, o := range []ClientOption{
		SetConnectionName("connection_name"),
		SetHost("host_value:8080"),
		SetVHost("vhost_value"),
		SetAuth("username_value", "pass_value"),
		SetTLSConfig(&tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12}),
		IsTLS(true),
		UseUnmarshaler(testUnmarshaler),
		UseMarshaler(defaultBytesMarshaler),
	} {
		o(&got)
	}

	want := newClientOptions()
	want.uri.Host = "host_value"
	want.uri.Vhost = "vhost_value"
	want.uri.Port = 8080
	want.uri.Username = "username_value"
	want.uri.Password = "pass_value"
	want.uri.Scheme = "amqps"
	want.config.Properties.SetClientConnectionName("connection_name")
	want.config.TLSClientConfig = &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12}
	want.unmarshaler[testUnmarshaler.ContentType()] = testUnmarshaler
	want.marshaler = defaultBytesMarshaler

	assert.Equal(t, want, got)
}

func TestClientOption_Validate(t *testing.T) {
	t.Parallel()

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		err := fmt.Errorf("error")
		got := (&clientOptions{err: err}).validate()
		assert.ErrorIs(t, got, err)
	})

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		got := newClientOptions()
		require.NoError(t, got.validate())
		assert.NotNil(t, got.logger)
		assert.NotNil(t, got.dialer)
	})
}

package amqpx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestPublisher_Reconnect(t *testing.T) {
	t.Parallel()

	client, mock := prep(t)
	defer client.Close()
	defer time.AfterFunc(defaultTimeout, func() { panic("deadlock") }).Stop()

	_ = NewPublisher[struct{}](client, Direct, UseRoutingKey("key"))
	done := make(chan bool)
	mock.Conn.ChannelFunc = func() (Channel, error) {
		defer close(done)
		return channelMock(), nil
	}
	mock.Conn.Close()
	<-done
}

func TestNewPublisher_BytesMarshaler(t *testing.T) {
	t.Parallel()

	client, _ := prep(t)
	defer client.Close()

	pub := NewPublisher[[]byte](client, Direct, UseRoutingKey("key"))
	assert.Equal(t, defaultBytesMarshaler, pub.marshaler)
}

func TestPublisher_Publish(t *testing.T) {
	t.Parallel()

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		client, mock := prep(t)
		defer client.Close()

		mock.Channel.PublishWithDeferredConfirmWithContextFunc = func(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp091.Publishing) (*amqp091.DeferredConfirmation, error) {
			assert.Equal(t, []byte("hello"), msg.Body)
			return nil, fmt.Errorf("failed")
		}

		pub := NewPublisher[[]byte](client, Direct, UseRoutingKey("key"))
		got := pub.Publish(NewPublishing([]byte("hello")))

		var err PublishError
		assert.ErrorAs(t, got, &err)
		assert.Equal(t, PublishError{Exchange: Direct, RoutingKey: "key", Message: "failed"}, err)
	})
}

func TestPublishing_Properties(t *testing.T) {
	t.Parallel()

	d := time.Now()
	got := NewPublishing([]byte(nil)).
		PersistentMode().
		SetPriority(1).
		SetCorrelationID("correlation_id_value").
		SetReplyTo("reply_to_value").
		SetExpiration("expiration_value").
		SetMessageID("message_id_value").
		SetTimestamp(d).
		SetType("type_value").
		SetUserID("user_id_value").
		SetAppID("app_id_value")
	got.WithContext(context.Background())

	want := &Publishing[[]byte]{
		req: &PublishRequest{
			Publishing: amqp091.Publishing{
				Headers:       amqp091.Table{},
				DeliveryMode:  Persistent,
				Priority:      1,
				CorrelationId: "correlation_id_value",
				ReplyTo:       "reply_to_value",
				Expiration:    "expiration_value",
				MessageId:     "message_id_value",
				Timestamp:     d,
				Type:          "type_value",
				UserId:        "user_id_value",
				AppId:         "app_id_value",
			},
			ctx: context.Background(),
		},
	}
	assert.Equal(t, want, got)
}

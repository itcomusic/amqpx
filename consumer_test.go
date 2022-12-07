package amqpx

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type unmarshaler struct{}

func (*unmarshaler) ContentType() string {
	return "application/json"
}

func (*unmarshaler) Unmarshal(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

var testUnmarshaler = &unmarshaler{}

func TestConsumer_Reconnect(t *testing.T) {
	t.Parallel()

	client, mock := prep(t)
	defer client.Close()
	defer time.AfterFunc(defaultTimeout, func() { panic("deadlock") }).Stop()

	assert.NoError(t, client.NewConsumer("", D(func(context.Context, *Delivery[[]byte]) Action { return Ack })))
	done := make(chan bool)
	mock.Conn.ChannelFunc = func() (Channel, error) {
		defer close(done)
		return channelMock(), nil
	}
	mock.Conn.Close()
	<-done
}

func TestClient_NewConsumer(t *testing.T) {
	t.Parallel()

	t.Run("channel error", func(t *testing.T) {
		t.Parallel()

		client, mock := prep(t)
		defer client.Close()

		mock.Conn.ChannelFunc = func() (Channel, error) {
			return nil, fmt.Errorf("failed")
		}

		got := client.NewConsumer("foo", D(func(context.Context, *Delivery[[]byte]) Action { return Ack }))
		var err ConsumerError
		assert.ErrorAs(t, got, &err)
		assert.Equal(t, ConsumerError{Queue: "foo", Message: "create channel: failed"}, err)
	})

	t.Run("consume error", func(t *testing.T) {
		t.Parallel()

		client, mock := prep(t)
		defer client.Close()

		mock.Channel.ConsumeFunc = func(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
			return nil, fmt.Errorf("failed")
		}

		got := client.NewConsumer("foo", D(func(context.Context, *Delivery[[]byte]) Action { return Ack }))
		var err ConsumerError
		assert.ErrorAs(t, got, &err)
		assert.Equal(t, ConsumerError{Queue: "foo", Message: "consume: failed"}, err)
	})
}

func TestDeliveryRequest_setStatus(t *testing.T) {
	t.Parallel()

	t.Run("ack", func(t *testing.T) {
		t.Parallel()

		ackMock := &AcknowledgerMock{
			AckFunc: func(tag uint64, multiple bool) error {
				return nil
			},
		}

		d := &DeliveryRequest{acknowledger: ackMock}
		require.NoError(t, d.setStatus(Ack))
		assert.Equal(t, 1, len(ackMock.AckCalls()))
	})

	t.Run("nack", func(t *testing.T) {
		t.Parallel()

		ackMock := &AcknowledgerMock{
			NackFunc: func(tag uint64, multiple bool, requeue bool) error {
				return nil
			},
		}

		d := &DeliveryRequest{acknowledger: ackMock}
		require.NoError(t, d.setStatus(Nack))
		assert.Equal(t, 1, len(ackMock.NackCalls()))
	})

	t.Run("reject", func(t *testing.T) {
		t.Parallel()

		ackMock := &AcknowledgerMock{
			RejectFunc: func(tag uint64, requeue bool) error {
				return nil
			},
		}

		d := &DeliveryRequest{acknowledger: ackMock}
		require.NoError(t, d.setStatus(Reject))
		assert.Equal(t, 1, len(ackMock.RejectCalls()))
	})
}

func TestDeliveryBytes(t *testing.T) {
	t.Parallel()

	client, mock := prep(t)
	defer client.Close()
	defer time.AfterFunc(defaultTimeout, func() { panic("deadlock") }).Stop()

	msg := amqp091.Delivery{
		Body: []byte("hello"),
		Acknowledger: &AcknowledgerMock{
			AckFunc: func(tag uint64, multiple bool) error {
				return nil
			},
		}}

	mock.Channel.ConsumeFunc = func(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
		ch := make(chan amqp091.Delivery, 1)
		ch <- msg
		return ch, nil
	}

	done := make(chan bool)
	got := client.NewConsumer("", D(func(ctx context.Context, d *Delivery[[]byte]) Action {
		defer close(done)
		assert.Equal(t, &msg.Body, d.Msg)
		return Ack
	}))
	require.NoError(t, got)
	<-done
}

func TestDeliveryStruct(t *testing.T) {
	t.Parallel()

	type Gopher struct {
		Name string `json:"name"`
	}

	var call bool
	fn := D[Gopher](func(ctx context.Context, got *Delivery[Gopher]) Action {
		call = true
		assert.Equal(t, &Delivery[Gopher]{Msg: &Gopher{Name: "gopher"}, Req: &DeliveryRequest{Body: nil, ContentType: testUnmarshaler.ContentType()}}, got)
		return Ack
	})
	fn.init(map[string]Unmarshaler{testUnmarshaler.ContentType(): testUnmarshaler})
	got := fn.serve(context.Background(), &DeliveryRequest{Body: []byte(`{"name":"gopher"}`), ContentType: testUnmarshaler.ContentType()})
	assert.Equal(t, true, call)
	assert.Equal(t, Ack, got)
}

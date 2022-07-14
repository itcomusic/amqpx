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
	defer time.AfterFunc(defaultTimeout, func() { t.Fatal("deadlock") }).Stop()

	assert.NoError(t, client.NewConsumer("", ConsumerFunc(func(*Delivery) Action { return Ack })))
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

		got := client.NewConsumer("", ConsumerFunc(func(*Delivery) Action { return Ack }))
		assert.EqualError(t, got, "amqpx: create channel: failed")
	})

	t.Run("consume error", func(t *testing.T) {
		t.Parallel()

		client, mock := prep(t)
		defer client.Close()

		mock.Channel.ConsumeFunc = func(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
			return nil, fmt.Errorf("failed")
		}

		got := client.NewConsumer("", ConsumerFunc(func(*Delivery) Action { return Ack }))
		assert.EqualError(t, got, "amqpx: consume: failed")
	})

	t.Run("unmarshaler error", func(t *testing.T) {
		t.Parallel()

		client, _ := prep(t)
		defer client.Close()

		client.unmarshaler = nil
		got := client.NewConsumer("", ConsumerFunc(func(*Delivery) Action { return Ack }))
		assert.ErrorIs(t, got, ErrUnmarshalerNotFound)
	})
}

func TestConsumer_AckMode(t *testing.T) {
	t.Parallel()

	t.Run("ack", func(t *testing.T) {
		t.Parallel()

		ackMock := &AcknowledgerMock{
			AckFunc: func(tag uint64, multiple bool) error {
				return nil
			},
		}

		d := &Delivery{
			Delivery: &amqp091.Delivery{
				Acknowledger: ackMock,
			},
		}
		d.setStatus(Ack)
		assert.Equal(t, 1, len(ackMock.AckCalls()))
	})

	t.Run("nack", func(t *testing.T) {
		t.Parallel()

		ackMock := &AcknowledgerMock{
			NackFunc: func(tag uint64, multiple bool, requeue bool) error {
				return nil
			},
		}

		d := &Delivery{
			Delivery: &amqp091.Delivery{
				Acknowledger: ackMock,
			},
		}
		d.setStatus(Nack)
		assert.Equal(t, 1, len(ackMock.NackCalls()))
	})

	t.Run("reject", func(t *testing.T) {
		t.Parallel()

		ackMock := &AcknowledgerMock{
			RejectFunc: func(tag uint64, requeue bool) error {
				return nil
			},
		}

		d := &Delivery{
			Delivery: &amqp091.Delivery{
				Acknowledger: ackMock,
			},
		}
		d.setStatus(Reject)
		assert.Equal(t, 1, len(ackMock.RejectCalls()))
	})
}

func TestConsumer_DeliveryBody(t *testing.T) {
	t.Parallel()

	client, mock := prep(t)
	defer client.Close()
	defer time.AfterFunc(defaultTimeout, func() { t.Error("deadlock") }).Stop()

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
	got := client.NewConsumer("", ConsumerFunc(func(d *Delivery) Action {
		defer close(done)
		assert.Equal(t, msg.Body, d.Body)
		return Ack
	}))
	require.NoError(t, got)
	<-done
}

func TestHandleValue_Serve(t *testing.T) {
	t.Parallel()

	type Gopher struct {
		Name string `json:"name"`
	}

	var call bool
	fn := ConsumerMessage[Gopher](func(ctx context.Context, m *Gopher) Action {
		call = true
		assert.Equal(t, &Gopher{Name: "gopher"}, m)
		return Ack
	})
	fn.init(map[string]Unmarshaler{testUnmarshaler.ContentType(): testUnmarshaler})
	got := fn.Serve(&Delivery{Delivery: &amqp091.Delivery{Body: []byte(`{"name":"gopher"}`), ContentType: testUnmarshaler.ContentType()}, ctx: context.Background()})
	assert.Equal(t, true, call)
	assert.Equal(t, Ack, got)
}

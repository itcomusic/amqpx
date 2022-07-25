package amqpx

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultTimeout = time.Second * 2

func TestClient_ConnectErr(t *testing.T) {
	t.Parallel()

	_, err := Connect(setDialer(&dialerMock{
		DialFunc: func(_ context.Context) (Connection, error) {
			return nil, fmt.Errorf("conn failed")
		},
	}))
	assert.EqualError(t, err, "conn failed")
}

func TestClient_Reconnect(t *testing.T) {
	t.Parallel()

	client, mock := prep(t)
	defer client.Close()

	done := make(chan bool, 1)
	mock.Dialer.DialFunc = func(_ context.Context) (Connection, error) {
		done <- true
		return mock.Conn, nil
	}
	mock.Conn.Close()

	select {
	case <-time.After(defaultTimeout):
		t.Errorf("no reconnect")
	case <-done:
	}
}

type mock struct {
	Dialer *dialerMock
	Conn   *ConnectionMock

	// returns the same channel for all
	Channel *ChannelMock
}

func channelMock() *ChannelMock {
	channel := &ChannelMock{
		QosFunc: func(prefetchCount int, prefetchSize int, global bool) error {
			return nil
		},
		ConsumeFunc: func(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error) {
			return make(chan amqp091.Delivery), nil
		},
		PublishWithDeferredConfirmWithContextFunc: func(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp091.Publishing) (*amqp091.DeferredConfirmation, error) {
			return nil, nil
		},
		NotifyCloseFunc: func(err chan *amqp091.Error) chan *amqp091.Error {
			return err
		},
		NotifyCancelFunc: func(tag chan string) chan string {
			return tag
		},
		NotifyReturnFunc: func(r chan amqp091.Return) chan amqp091.Return {
			return r
		},
	}

	m := sync.Mutex{}
	channel.CloseFunc = func() error {
		m.Lock()
		defer m.Unlock()

		for _, v := range channel.NotifyCloseCalls() {
			select {
			case <-v.ErrorCh:
			default:
				close(v.ErrorCh)
			}
		}
		return nil
	}
	return channel
}

func prep(t *testing.T) (*Client, mock) {
	channel := channelMock()
	conn := &ConnectionMock{
		IsClosedFunc: func() bool {
			return false
		},
		NotifyCloseFunc: func(errorCh chan *amqp091.Error) chan *amqp091.Error {
			return errorCh
		},
		ChannelFunc: func() (Channel, error) {
			return channel, nil
		},
	}

	conn.CloseFunc = func() error {
		defer channel.Close()
		for _, v := range conn.NotifyCloseCalls() {
			select {
			case <-v.ErrorCh:
			default:
				close(v.ErrorCh)
			}
		}
		return nil
	}

	mock := mock{
		Dialer: &dialerMock{
			DialFunc: func(_ context.Context) (Connection, error) {
				return conn, nil
			},
		},
		Conn:    conn,
		Channel: channel,
	}

	client, err := Connect(setDialer(mock.Dialer),
		WithLogger(func(err error) { t.Fatal(err) }),
		UseUnmarshaler(testUnmarshaler),
		UseMarshaler(defaultBytesMarshaler))
	require.NoError(t, err)
	return client, mock
}

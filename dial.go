package amqpx

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	reconnectDelay = time.Second * 4
)

//go:generate ./bin/moq -rm -out connection_moq_test.go . Connection

// A Connection is an interface implemented by amqp091 client.
type Connection interface {
	IsClosed() bool
	Channel() (Channel, error)
	NotifyClose(chan *amqp091.Error) chan *amqp091.Error
	Close() error
}

//go:generate ./bin/moq -rm -out channel_moq_test.go . Channel

// A Channel is an interface implemented by amqp091 client.
type Channel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp091.Table) (amqp091.Queue, error)
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error
	QueueBind(name, key, exchange string, noWait bool, args amqp091.Table) error
	Confirm(noWait bool) error
	Qos(prefetchCount, prefetchSize int, global bool) error
	NotifyClose(chan *amqp091.Error) chan *amqp091.Error
	NotifyCancel(chan string) chan string
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp091.Table) (<-chan amqp091.Delivery, error)
	PublishWithDeferredConfirmWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) (*amqp091.DeferredConfirmation, error)
	NotifyReturn(c chan amqp091.Return) chan amqp091.Return
	Close() error
}

//go:generate ./bin/moq -out dialer_moq_test.go . dialer

// A dialer is an interface implemented by creating connection.
type dialer interface {
	Dial(context.Context) (Connection, error)
}

type amqpConn struct {
	conn *amqp091.Connection
}

func (w *amqpConn) IsClosed() bool {
	return w.conn.IsClosed()
}

func (w *amqpConn) Channel() (Channel, error) {
	return w.conn.Channel()
}

func (w *amqpConn) NotifyClose(receiver chan *amqp091.Error) chan *amqp091.Error {
	return w.conn.NotifyClose(receiver)
}

func (w *amqpConn) Close() error {
	return w.conn.Close()
}

var defaultURI = amqp091.URI{
	Scheme:   "amqp",
	Host:     "localhost",
	Port:     5672,
	Username: "guest",
	Password: "guest",
	Vhost:    "/",
}

var defaultConfig = amqp091.Config{
	Heartbeat:  10 * time.Second,
	Locale:     "en_US",
	Properties: amqp091.NewConnectionProperties(),
}

type defaultDialer struct {
	URI     string
	Config  amqp091.Config
	logFunc LogFunc
}

func (d *defaultDialer) Dial(ctx context.Context) (Connection, error) {
	for {
		conn, err := amqp091.DialConfig(d.URI, d.Config)
		if err == nil {
			return &amqpConn{conn: conn}, nil
		}
		d.logFunc(fmt.Errorf("dial conn: %w", err))

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%s: %w", err, ctx.Err())

		case <-time.After(reconnectDelay):
		}
	}
}

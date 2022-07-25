// Package amqpx provides working with RabbitMQ using AMQP 0.9.1.
package amqpx

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

var (
	errChannelClosed       = fmt.Errorf("channel/connection is not open")
	errPublishConfirm      = fmt.Errorf("publish has not confirmation")
	errUnmarshalerNotFound = fmt.Errorf("unmarshaler not found")
	errMarshalerNotFound   = fmt.Errorf("marshaler not found")
	errRoutingKeyEmpty     = fmt.Errorf("routing-key is empty")

	errConnClosed = fmt.Errorf("connection closed")
	errFuncNil    = fmt.Errorf("consumer func nil")
)

// The delivery mode of messages is unrelated to the durability of the queues they reside on.
const (
	// Transient means higher throughput but messages will not be restored on broker restart.
	// Transient messages will not be restored to durable queues.
	Transient = amqp091.Transient

	// Persistent messages will be restored to
	// durable queues and lost on non-durable queues during server restart.
	Persistent = amqp091.Persistent
)

type Table = amqp091.Table

// Default exchanges.
const (
	Direct  = "amq.direct"
	Fanout  = "amq.fanout"
	Headers = "amq.headers"
	Match   = "amq.match"
	Topic   = "amq.topic"
)

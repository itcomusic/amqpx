package amqpx

import (
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

//go:generate ./bin/stringer -type=Action

// Action represents acknowledgment status the delivered message.
type Action int8

const (
	// Ack is acknowledgement that the client or server has finished work on a delivery.
	// It removes message from the queue permanently.
	Ack Action = iota

	// Nack is a negatively acknowledge the delivery of message and need requeue.
	//
	// The server to deliver this message to a different consumer.
	// If it is not possible the message will be dropped or delivered to a server configured dead-letter queue.
	//
	// This action must not be used to select or requeue messages the client wishes
	// not to handle, rather it is to inform the server that the client is incapable
	// of handling this message at this time.
	Nack

	// Reject is an explicit not acknowledged and do not requeue.
	Reject
)

type DeliveryRequest struct {
	Headers Table // Application or header exchange table

	// Properties
	ContentType     string    // MIME content type
	ContentEncoding string    // MIME content encoding
	DeliveryMode    uint8     // queue implementation use - non-persistent (1) or persistent (2)
	Priority        uint8     // queue implementation use - 0 to 9
	CorrelationID   string    // application use - correlation identifier
	ReplyTo         string    // application use - address to reply to (ex: RPC)
	Expiration      string    // implementation use - message expiration spec
	MessageID       string    // application use - message identifier
	Timestamp       time.Time // application use - message timestamp
	Type            string    // application use - message type name
	UserID          string    // application use - creating user - should be authenticated user
	AppID           string    // application use - creating application id
	ConsumerTag     string
	DeliveryTag     uint64
	Redelivered     bool
	Exchange        string // basic.publish exchange
	RoutingKey      string // basic.publish routing key
	Body            []byte

	status       Action
	acknowledger Acknowledger // the channel from which this delivery arrived
	log          LogFunc
}

func newDeliveryRequest(d *amqp091.Delivery, l LogFunc) *DeliveryRequest {
	if d.Headers == nil {
		d.Headers = make(amqp091.Table)
	}

	return &DeliveryRequest{
		Headers:         d.Headers,
		ContentType:     d.ContentType,
		ContentEncoding: d.ContentEncoding,
		DeliveryMode:    d.DeliveryMode,
		Priority:        d.Priority,
		CorrelationID:   d.CorrelationId,
		ReplyTo:         d.ReplyTo,
		Expiration:      d.Expiration,
		MessageID:       d.MessageId,
		Timestamp:       d.Timestamp,
		Type:            d.Type,
		UserID:          d.UserId,
		AppID:           d.AppId,
		ConsumerTag:     d.ConsumerTag,
		DeliveryTag:     d.DeliveryTag,
		Redelivered:     d.Redelivered,
		Exchange:        d.Exchange,
		RoutingKey:      d.RoutingKey,
		Body:            d.Body,
		acknowledger:    d.Acknowledger,
		log:             l,
	}
}

// Status returns acknowledgement status.
func (d *DeliveryRequest) Status() Action {
	return d.status
}

func (d *DeliveryRequest) Log(format string, v ...any) {
	d.log(format, v...)
}

func (d *DeliveryRequest) setStatus(status Action) error {
	switch status {
	case Ack:
		return d.ack()

	case Nack:
		return d.nack()

	case Reject:
		return d.reject()

	default:
		return fmt.Errorf("delivery has unknown ack mode \"%d\"", status)
	}
}

func (d *DeliveryRequest) ack() error {
	if err := d.acknowledger.Ack(d.DeliveryTag, false); err != nil {
		return err
	}

	d.status = Ack
	return nil
}

func (d *DeliveryRequest) nack() error {
	if err := d.acknowledger.Nack(d.DeliveryTag, false, true); err != nil {
		return err
	}

	d.status = Nack
	return nil
}

func (d *DeliveryRequest) reject() error {
	if err := d.acknowledger.Reject(d.DeliveryTag, false); err != nil {
		return err
	}

	d.status = Reject
	return nil
}

func (d *DeliveryRequest) info() string {
	return fmt.Sprintf("exchange %q routing-key %q content-type %q", d.Exchange, d.RoutingKey, d.ContentType)
}

// A Delivery represent the fields for a delivered message.
type Delivery[T any] struct {
	Msg *T
	Req *DeliveryRequest
}

func newDelivery[T any](v *T, req *DeliveryRequest) *Delivery[T] {
	return &Delivery[T]{
		Msg: v,
		Req: req,
	}
}

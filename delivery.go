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
	// Ack is an acknowledgement that the client or server has finished work on a delivery.
	// It removes a message from the queue permanently.
	Ack Action = iota

	// Nack is a negatively acknowledging the delivery of message and need requeue.
	//
	// The server to deliver this message to a different consumer.
	// If it is not possible, the message will be dropped or delivered to a server configured dead-letter queue.
	//
	// This action must not be used to select or re-queue messages the client wishes
	// not to handle, rather it is to inform the server that the client is incapable
	// of handling this message at this time.
	Nack

	// Reject is explicit not acknowledged and do not requeue.
	Reject
)

type DeliveryRequest struct {
	in     *amqp091.Delivery
	status Action
	log    LogFunc
}

func newDeliveryRequest(req *amqp091.Delivery, l LogFunc) *DeliveryRequest {
	if req.Headers == nil {
		req.Headers = make(amqp091.Table)
	}

	return &DeliveryRequest{
		in:  req,
		log: l,
	}
}

// NewFrom using from only tests.
func (d *DeliveryRequest) NewFrom(req *amqp091.Delivery) *DeliveryRequest {
	return &DeliveryRequest{in: req}
}

// A Headers returns the headers of the message.
func (d *DeliveryRequest) Headers() Table {
	return d.in.Headers
}

// A ContentType returns the content type of the message.
func (d *DeliveryRequest) ContentType() string {
	return d.in.ContentType
}

// A ContentEncoding returns the content encoding of the message.
func (d *DeliveryRequest) ContentEncoding() string {
	return d.in.ContentEncoding
}

// A DeliveryMode returns the delivery mode of the message.
func (d *DeliveryRequest) DeliveryMode() uint8 {
	return d.in.DeliveryMode
}

// A Priority returns the priority of the message.
func (d *DeliveryRequest) Priority() uint8 {
	return d.in.Priority
}

// A CorrelationID returns the correlation identifier of the message.
func (d *DeliveryRequest) CorrelationID() string {
	return d.in.CorrelationId
}

// A ReplyTo returns the address to reply to (ex: RPC).
func (d *DeliveryRequest) ReplyTo() string {
	return d.in.ReplyTo
}

// An Expiration returns the expiration of the message.
func (d *DeliveryRequest) Expiration() string {
	return d.in.Expiration
}

// A MessageID returns the application message identifier.
func (d *DeliveryRequest) MessageID() string {
	return d.in.MessageId
}

// A Timestamp returns the message timestamp.
func (d *DeliveryRequest) Timestamp() time.Time {
	return d.in.Timestamp
}

// A Type returns the message type name.
func (d *DeliveryRequest) Type() string {
	return d.in.Type
}

// A UserID returns the creating user id.
func (d *DeliveryRequest) UserID() string {
	return d.in.UserId
}

// An AppID returns the creating application id.
func (d *DeliveryRequest) AppID() string {
	return d.in.AppId
}

// A ConsumerTag returns the consumer tag.
func (d *DeliveryRequest) ConsumerTag() string {
	return d.in.ConsumerTag
}

// A DeliveryTag returns the server-assigned delivery tag.
func (d *DeliveryRequest) DeliveryTag() uint64 {
	return d.in.DeliveryTag
}

// A Redelivered returns whether this is a redelivery of a message.
func (d *DeliveryRequest) Redelivered() bool {
	return d.in.Redelivered
}

// An Exchange returns the exchange name.
func (d *DeliveryRequest) Exchange() string {
	return d.in.Exchange
}

// A RoutingKey returns the routing key.
func (d *DeliveryRequest) RoutingKey() string {
	return d.in.RoutingKey
}

// A Body returns the body of the message.
func (d *DeliveryRequest) Body() []byte {
	return d.in.Body
}

// SetBody sets the body of the message.
func (d *DeliveryRequest) SetBody(b []byte) {
	d.in.Body = b
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
	if err := d.in.Acknowledger.Ack(d.in.DeliveryTag, false); err != nil {
		return err
	}

	d.status = Ack
	return nil
}

func (d *DeliveryRequest) nack() error {
	if err := d.in.Acknowledger.Nack(d.in.DeliveryTag, false, true); err != nil {
		return err
	}

	d.status = Nack
	return nil
}

func (d *DeliveryRequest) reject() error {
	if err := d.in.Acknowledger.Reject(d.in.DeliveryTag, false); err != nil {
		return err
	}

	d.status = Reject
	return nil
}

func (d *DeliveryRequest) info() string {
	return fmt.Sprintf("exchange %q routing-key %q content-type %q", d.in.Exchange, d.in.RoutingKey, d.in.ContentType)
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

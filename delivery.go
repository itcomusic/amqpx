package amqpx

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

//go:generate ./bin/stringer -type=Action

// Action represents acknowledgment status the delivered message.
type Action int8

const (
	Ack Action = iota
	Nack
	Reject
)

// A Delivery represent the fields for a delivered message.
type Delivery struct {
	*amqp091.Delivery

	logFunc LogFunc
	status  Action
	ctx     context.Context
}

func newDelivery(d *amqp091.Delivery, log LogFunc) *Delivery {
	if d.Headers == nil {
		d.Headers = make(amqp091.Table)
	}

	return &Delivery{
		Delivery: d,
		logFunc:  log,
		ctx:      context.Background(),
	}
}

// Status returns acknowledgement status.
func (d *Delivery) Status() Action {
	return d.status
}

// Context returns context.
func (d *Delivery) Context() context.Context {
	return d.ctx
}

// WithContext sets context.
func (d *Delivery) WithContext(ctx context.Context) {
	d.ctx = ctx
}

func (d *Delivery) setStatus(status Action) {
	switch status {
	case Ack:
		d.ack()

	case Nack:
		d.nack()

	case Reject:
		d.reject()

	default:
		d.logFunc("[ERROR] delivery has unknown ack mode: %d", status)
	}
}

// ack is acknowledgement that the client or server has finished work on a delivery.
// It removes message from the queue permanently.
func (d *Delivery) ack() {
	if err := d.Delivery.Ack(false); err != nil {
		d.logFunc("[ERROR] ack: %s", err)
		return
	}
	d.status = Ack
}

// nack is a negatively acknowledge the delivery of message and need requeue (by default).
//
// When requeue is true request the server to deliver this message to a different
// consumer. If it is not possible or requeue is false, the message will be
// dropped or delivered to a server configured dead-letter queue.
//
// This method must not be used to select or requeue messages the client wishes
// not to handle, rather it is to inform the server that the client is incapable
// of handling this message at this time.
func (d *Delivery) nack() {
	if err := d.Delivery.Nack(false, true); err != nil {
		d.logFunc("[ERROR] nack: %s", err)
		return
	}
	d.status = Nack
}

// reject is an explicit not acknowledged and do not requeue (by default).
// When requeue is true, queue this message to be delivered to a consumer on a different channel.
func (d *Delivery) reject() {
	if err := d.Delivery.Reject(false); err != nil {
		d.logFunc("[ERROR] reject: %s", err)
		return
	}
	d.status = Reject
}

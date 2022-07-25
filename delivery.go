package amqpx

import (
	"context"
	"fmt"

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

func (d *Delivery) setStatus(status Action) error {
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

func (d *Delivery) ack() error {
	if err := d.Delivery.Ack(false); err != nil {
		return err
	}

	d.status = Ack
	return nil
}

func (d *Delivery) nack() error {
	if err := d.Delivery.Nack(false, true); err != nil {
		return err
	}
	d.status = Nack
	return nil
}

func (d *Delivery) reject() error {
	if err := d.Delivery.Reject(false); err != nil {
		return err
	}
	d.status = Reject
	return nil
}

package amqpx

import (
	"fmt"
)

type ConsumerError struct {
	Queue   string
	Tag     string
	Message string
}

func (c ConsumerError) Error() string {
	return fmt.Sprintf("amqpx: queue %q consumer-tag %q: %s", c.Queue, c.Tag, c.Message)
}

type DeliveryError struct {
	Exchange    string
	RoutingKey  string
	ContentType string
	Message     string
}

func (m DeliveryError) Error() string {
	return fmt.Sprintf("amqpx: exchange %q routing-key %q content-type %q: %s", m.Exchange, m.RoutingKey, m.ContentType, m.Message)
}

type PublishError struct {
	Exchange   string
	RoutingKey string
	Message    string
}

func (p PublishError) Error() string {
	return fmt.Sprintf("amqpx: exchange %q routing-key %q: %s", p.Exchange, p.RoutingKey, p.Message)
}

type ReturnError struct {
	Exchange   string
	RoutingKey string
	ReplyText  string
	ReplyCode  uint16
}

func (r ReturnError) Error() string {
	return fmt.Sprintf("amqpx: exchange %q routing-key %q undeliverable message desc %q \"%d\"", r.Exchange, r.RoutingKey, r.ReplyText, r.ReplyCode)
}

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
	return fmt.Sprintf("amqpx: queue \"%s\" consumer-tag \"%s\": %s", c.Queue, c.Tag, c.Message)
}

type DeliveryError struct {
	Exchange    string
	RoutingKey  string
	ContentType string
	Message     string
}

func (m DeliveryError) Error() string {
	return fmt.Sprintf("amqpx: exchange \"%s\" routing-key \"%s\" content-type \"%s\": %s", m.Exchange, m.RoutingKey, m.ContentType, m.Message)
}

type PublisherError struct {
	Exchange   string
	RoutingKey string
	Message    string
}

func (p PublisherError) Error() string {
	return fmt.Sprintf("amqpx: exchange \"%s\" routing-key \"%s\": %s", p.Exchange, p.RoutingKey, p.Message)
}

type ReturnError struct {
	Exchange   string
	RoutingKey string
	ReplyText  string
	ReplyCode  uint16
}

func (r ReturnError) Error() string {
	return fmt.Sprintf("amqpx: exchange \"%s\" routing-key \"%s\" undeliverable message desc \"%s\" \"%d\"", r.Exchange, r.RoutingKey, r.ReplyText, r.ReplyCode)
}

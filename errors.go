package amqpx

import "fmt"

type ConsumerError struct {
	Queue   string
	Tag     string
	Message string
}

func (c ConsumerError) Error() string {
	return fmt.Sprintf("amqpx: queue \"%s\" consumer-tag \"%s\": %s", c.Queue, c.Tag, c.Message)
}

type MessageError struct {
	Exchange   string
	RoutingKey string
	Message    string
}

func (m MessageError) Error() string {
	return fmt.Sprintf("amqpx: exchange \"%s\" routing-key \"%s\": %s", m.Exchange, m.RoutingKey, m.Message)
}

type PublisherError struct {
	Exchange   string
	RoutingKey string
	Message    string
}

func (p PublisherError) Error() string {
	return fmt.Sprintf("amqpx: exchange \"%s\" routing-key \"%s\": %s", p.Exchange, p.RoutingKey, p.Message)
}

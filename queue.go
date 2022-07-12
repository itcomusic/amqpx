package amqpx

// A QueueDeclare represents a queue declaration.
type QueueDeclare struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       Table
}

// A ExchangeDeclare represents an exchange declaration.
type ExchangeDeclare struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       Table
}

// A QueueBind represents a binding routing key, exchange and queue declaration.
type QueueBind struct {
	Exchange   string
	RoutingKey []string
	NoWait     bool
	Args       Table
}

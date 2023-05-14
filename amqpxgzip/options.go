package amqpxgzip

type config struct {
	level int
}

// An Option configures the GZIP.
type Option func(*config)

// WithCompression configures the supplied compression level.
func WithCompression(level int) Option {
	return func(c *config) {
		if level != 0 {
			c.level = level
		}
	}
}

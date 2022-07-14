package amqpx

import (
	"log"
	"os"
)

// LogFunc type is an adapter to allow the use of ordinary functions as LogFunc.
type LogFunc func(format string, args ...any)

// NoopLogger logFunc does nothing
var NoopLogger = LogFunc(func(_ string, _ ...any) {})

var defaultLogger = func() LogFunc {
	logger := log.New(os.Stderr, "amqpx: ", log.LstdFlags)
	return func(format string, args ...any) {
		logger.Printf(format, args...)
	}
}()

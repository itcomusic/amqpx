package amqpx

import (
	"log"
	"os"
)

// LogFunc type is an adapter to allow the use of ordinary functions as LogFunc.
type LogFunc func(format string, v ...any)

// NoOpLogger logger does nothing
var NoOpLogger = LogFunc(func(format string, v ...any) {})

var defaultLogger = func() LogFunc {
	l := log.New(os.Stderr, "amqpx: ", log.LstdFlags)
	return func(format string, v ...any) {
		l.Printf(format, v...)
	}
}()

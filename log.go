package amqpx

import (
	"log"
	"os"
)

// LogFunc type is an adapter to allow the use of ordinary functions as LogFunc.
type LogFunc func(error)

// NoOpLogger logger does nothing
var NoOpLogger = LogFunc(func(error) {})

var defaultLogger = func() LogFunc {
	logger := log.New(os.Stderr, "amqpx: ", log.LstdFlags)
	return func(err error) {
		logger.Println(err.Error())
	}
}()

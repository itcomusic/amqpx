package amqpx

// LogFunc type is an adapter to allow the use of ordinary functions as LogFunc.
type LogFunc func(format string, args ...any)

// NoopLogger logFunc does nothing
var NoopLogger = LogFunc(func(_ string, _ ...any) {})

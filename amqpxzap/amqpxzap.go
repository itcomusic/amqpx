// Package amqpxzap provides logs using the zap.
package amqpxzap

import (
	"go.uber.org/zap"

	"github.com/itcomusic/amqpx"
)

// S returns amqpx.LogFunc that writes to the zap logger at info-level.
func S(l *zap.SugaredLogger) amqpx.LogFunc {
	return func(err error) {
		l.Errorw("amqpx", "error", err)
	}
}

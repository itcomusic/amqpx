// Package amqpxgzip provides supporting gzip.
package amqpxgzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/itcomusic/amqpx"
)

const headerGZIP = "gzip"

// Consumer returns consume hook that wraps the next.
func Consumer() amqpx.ConsumeHook {
	return func(next amqpx.Consume) amqpx.Consume {
		return amqpx.D(func(d *amqpx.Delivery) amqpx.Action {
			if d.ContentEncoding != headerGZIP {
				return next.Serve(d)
			}

			r, err := gzip.NewReader(bytes.NewReader(d.Body))
			if err != nil {
				d.Log(fmt.Errorf("amqpxgzip: init reader: %w", err))
				return amqpx.Reject
			}
			defer r.Close()

			b, err := io.ReadAll(r)
			if err != nil {
				d.Log(fmt.Errorf("amqpxgzip: read: %w", err))
				return amqpx.Reject
			}

			d.Body = b
			return next.Serve(d)
		})
	}
}

// Publisher returns publish hook that wraps the next.
func Publisher(level ...int) amqpx.PublishHook {
	compression := gzip.DefaultCompression
	if len(level) != 0 {
		compression = level[0]
	}

	return func(next amqpx.PublisherFunc) amqpx.PublisherFunc {
		return func(m *amqpx.Publishing) error {
			buf := &bytes.Buffer{}
			w, err := gzip.NewWriterLevel(buf, compression)
			if err != nil {
				return fmt.Errorf("amqpxgzip: init writer: %w", err)
			}

			if _, err := w.Write(m.Body); err != nil {
				return fmt.Errorf("amqpxgzip: write: %w", err)
			}

			if err := w.Close(); err != nil {
				return fmt.Errorf("amqpxgzip: close: %w", err)
			}

			m.ContentEncoding = headerGZIP
			m.Body = buf.Bytes()
			return next(m)
		}
	}
}

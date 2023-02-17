// Package amqpxgzip provides supporting gzip.
package amqpxgzip

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/itcomusic/amqpx"
)

const headerGZIP = "gzip"

// Consumer returns consume hook that wraps the next.
func Consumer() amqpx.ConsumeHook {
	return func(next amqpx.ConsumeFunc) amqpx.ConsumeFunc {
		return func(ctx context.Context, req *amqpx.DeliveryRequest) amqpx.Action {
			if req.ContentEncoding != headerGZIP {
				return next(ctx, req)
			}

			r, err := gzip.NewReader(bytes.NewReader(req.Body))
			if err != nil {
				req.Log("[ERROR] amqpxgzip: init reader: %s", err)
				return amqpx.Reject
			}
			defer r.Close()

			body, err := io.ReadAll(r)
			if err != nil {
				req.Log("[ERROR] amqpxgzip: read: %s", err)
				return amqpx.Reject
			}

			req.Body = body
			return next(ctx, req)
		}
	}
}

// Publisher returns publish hook that wraps the next.
func Publisher(level ...int) amqpx.PublishHook {
	compression := gzip.DefaultCompression
	if len(level) != 0 {
		compression = level[0]
	}

	return func(next amqpx.PublisherFunc) amqpx.PublisherFunc {
		return func(ctx context.Context, m *amqpx.PublishRequest) error {
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
			return next(ctx, m)
		}
	}
}

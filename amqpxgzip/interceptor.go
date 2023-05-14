package amqpxgzip

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/itcomusic/amqpx"
)

type Interceptor struct {
	config config
}

var _ amqpx.Interceptor = (*Interceptor)(nil)

// NewInterceptor returns a new gzip interceptor.
func NewInterceptor(opts ...Option) *Interceptor {
	cfg := config{level: gzip.DefaultCompression}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Interceptor{
		config: cfg,
	}
}

func (i *Interceptor) WrapConsume(next amqpx.ConsumeFunc) amqpx.ConsumeFunc {
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

func (i *Interceptor) WrapPublish(next amqpx.PublishFunc) amqpx.PublishFunc {
	return func(ctx context.Context, m *amqpx.PublishingRequest) error {
		buf := &bytes.Buffer{}
		w, err := gzip.NewWriterLevel(buf, i.config.level)
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

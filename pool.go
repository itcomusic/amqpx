package amqpx

import (
	"sync"
)

// A pool is a generic wrapper around a sync.Pool.
type pool[T any] struct {
	pool  sync.Pool
	reset func(*T)
}

func (p *pool[T]) Get() *T {
	if p.reset != nil {
		return p.pool.Get().(*T)
	}
	return new(T)
}

func (p *pool[T]) Put(v *T) {
	if p.reset == nil {
		return
	}
	p.reset(v)
	p.pool.Put(v)
}

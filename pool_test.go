package amqpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_DisableSyncPool(t *testing.T) {
	pool := &pool[string]{}
	obj1 := pool.Get()
	v := "Gopher"
	*obj1 = v

	pool.Put(obj1)
	assert.NotNil(t, obj1)

	obj2 := pool.Get()
	assert.NotEqual(t, obj1, obj2)
}

func TestPool_EnableSyncPool(t *testing.T) {
	type Gopher struct {
		Name *string
	}

	pool := &pool[Gopher]{}
	SetPool[Gopher](func(v *Gopher) { v.Name = nil })(pool)
	obj1 := pool.Get()
	v := "Gopher"
	obj1.Name = &v

	pool.Put(obj1)
	require.Nil(t, obj1.Name)
	// error using sync pool, out from scope
	obj1.Name = &v

	obj2 := pool.Get()
	assert.Equal(t, obj1, obj2)
}

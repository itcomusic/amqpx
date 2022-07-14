package amqpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_DisableSyncPool(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	type Gopher struct {
		Name *string
	}

	pool := &pool[Gopher]{}
	SetPool[Gopher](func(v *Gopher) { v.Name = nil })(pool)

	got1 := pool.Get()
	v := "Gopher"
	got1.Name = &v

	pool.Put(got1)
	require.Nil(t, got1.Name)

	got2 := pool.Get()
	assert.Equal(t, got1, got2)
}

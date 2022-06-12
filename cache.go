package gbind

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type cache struct {
	lock sync.Mutex
	m    atomic.Value
}

func newCache() *cache {
	c := &cache{}
	c.m.Store(make(map[reflect.Type]*structType))
	return c
}

func (c *cache) get(key reflect.Type) (g *structType, found bool) {
	g, found = c.m.Load().(map[reflect.Type]*structType)[key]
	return
}

func (c *cache) set(key reflect.Type, value *structType) {
	c.lock.Lock()
	defer c.lock.Unlock()
	m := c.m.Load().(map[reflect.Type]*structType)
	nm := make(map[reflect.Type]*structType, len(m)+1)
	for k, v := range m {
		nm[k] = v
	}
	nm[key] = value
	c.m.Store(nm)
}

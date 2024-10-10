package local

import (
	"context"
	"github.com/ecodeclub/ekit/syncx"
	"interview-cases/case21_30/case24/repository/cache"
)


type Cache struct {
	kvs syncx.Map[string, any]
}

func (c *Cache) Get(ctx context.Context,key string) (any, error) {
	v,ok := c.kvs.Load(key)
	if !ok  {
		return nil, cache.ErrKeyNotFound
	}
	return v, nil
}

func (c *Cache) Set(ctx context.Context,key string, val any) error {
	c.kvs.Store(key, val)
	return nil
}

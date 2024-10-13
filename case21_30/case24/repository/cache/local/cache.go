package local

import (
	"context"
	"github.com/ecodeclub/ekit/syncx"
	"interview-cases/case21_30/case24/domain"
	"interview-cases/case21_30/case24/repository/cache"
)

type Cache struct {
	kvs syncx.Map[int64, domain.Order]
}

func NewCache()*Cache {
	return &Cache{
		kvs: syncx.Map[int64, domain.Order]{},
	}
}

func (c *Cache) Get(ctx context.Context, id int64) (domain.Order, error) {
	v, ok := c.kvs.Load(id)
	if !ok {
		return domain.Order{}, cache.ErrKeyNotFound
	}
	return v, nil
}

func (c *Cache) Set(ctx context.Context, order domain.Order) error {
	c.kvs.Store(order.ID, order)
	return nil
}

func (c *Cache) Del(ctx context.Context, id int64) error {
	c.kvs.Delete(id)
	return nil
}

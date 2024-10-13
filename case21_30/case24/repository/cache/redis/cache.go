package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"interview-cases/case21_30/case24/domain"
	"interview-cases/case21_30/case24/repository/cache"
)

type Cache struct {
	client redis.Cmdable
}

func NewCache(client redis.Cmdable) *Cache {
	return &Cache{
		client: client,
	}
}

func (c *Cache) key(id int64) string {
	return fmt.Sprintf("order:%d", id)
}

func (c *Cache) Get(ctx context.Context, id int64) (domain.Order, error) {
	val, err := c.client.Get(ctx, c.key(id)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return domain.Order{}, cache.ErrKeyNotFound
		}
		return domain.Order{}, err
	}
	var order domain.Order
	err = json.Unmarshal([]byte(val), &order)
	if err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

func (c *Cache) Set(ctx context.Context, val domain.Order) error {
	orderStr, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.key(val.ID), string(orderStr), 0).Err()
}
func (c *Cache) Del(ctx context.Context, id int64) error {
	return c.client.Del(ctx, c.key(id)).Err()
}

func (c *Cache) Ping(ctx context.Context) error {
	res, err := c.client.Ping(ctx).Result()
	if err != nil {
		return err
	}
	if res != "PONG" {
		return errors.New("ping不通")
	}
	return nil
}

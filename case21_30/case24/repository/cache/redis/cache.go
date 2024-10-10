package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client redis.Cmdable
}

func (c *Cache) Get(ctx context.Context, key string) (any, error) {
	val, err := c.client.Get(ctx, key).Result()
	return val, err
}

func (c *Cache) Set(ctx context.Context, key string, val any) error {
	return c.client.Set(ctx, key, val, 0).Err()
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

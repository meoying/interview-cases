package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	main   *redis.Client
	backup *redis.Client
}

func NewRedis(main *redis.Client, backup *redis.Client) *Redis {
	return &Redis{
		main:   main,
		backup: backup,
	}
}

func (r *Redis) Set(ctx context.Context, key string, val any) error {
	re := r.getRedis(ctx)
	if err := re.Set(ctx, key, val, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	re := r.getRedis(ctx)
	res, err := re.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return res, nil
}

func (r *Redis) getRedis(ctx context.Context) *redis.Client {
	if err := r.main.Ping(ctx).Err(); err == nil {
		return r.main
	} else {
		return r.backup
	}
}

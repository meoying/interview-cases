package test

import "github.com/redis/go-redis/v9"

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func InitRedisBloom() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})
}

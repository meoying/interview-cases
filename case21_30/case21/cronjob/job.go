package cronjob

import (
	"context"
	"interview-cases/case21_30/case21/repository/cache/local"
	"interview-cases/case21_30/case21/repository/cache/redis"
	"log/slog"

	"time"
)

// ToRedisJob 同步到redis
type ToRedisJob struct {
	redisCache *redis.Cache
	triSvc     TriSvc
}

func NewDBToRedisJob(redisCache *redis.Cache, triSvc TriSvc) *ToRedisJob {
	return &ToRedisJob{
		redisCache: redisCache,
		triSvc:     triSvc,
	}
}

func (d *ToRedisJob) Run() {
	// 从db取出1000条数据
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	items, err := d.triSvc.TopN(ctx, 1000)
	if err != nil {
		// 记录一下错误日志
		slog.Error("从全局获取获取数据失败", slog.Any("err", err))
		return
	}
	// 记录到redis
	err = d.redisCache.Set(ctx, items)
	if err != nil {
		// 记录一下错误日志
		slog.Error("数据同步到redis失败", slog.Any("err", err))
		return
	}
}

// RedisToLocalJob redis同步到本地缓存
type RedisToLocalJob struct {
	redisCache *redis.Cache
	localCache *local.Cache
}

func NewRedisToLocalJob(redisCache *redis.Cache, localCache *local.Cache) *RedisToLocalJob {
	return &RedisToLocalJob{
		redisCache: redisCache,
		localCache: localCache,
	}
}

func (r *RedisToLocalJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 获取100条
	items, err := r.redisCache.Get(ctx, 99)
	if err != nil {
		// 记录一下错误日志
		slog.Error("从redis获取数据失败", slog.Any("err", err))
		return
	}
	// 刷新到本地缓存
	err = r.localCache.Set(ctx, items)
	if err != nil {
		slog.Error("同步数据到本地失败", slog.Any("err", err))
	}
}

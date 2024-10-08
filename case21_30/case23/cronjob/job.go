package cronjob

import (
	"context"
	"interview-cases/case21_30/case23/repository/cache/local"
	"interview-cases/case21_30/case23/repository/cache/redis"
	"log/slog"
	"time"
)

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
	items, err := r.redisCache.Get(ctx, 100)
	if err != nil {
		// 记录一下错误日志
		slog.Error("从redis获取数据失败", slog.Any("err", err))
		return
	}
	// 刷新到本地缓存
	err = r.localCache.Set(ctx, items)
}

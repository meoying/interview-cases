package mix

import (
	"context"
	"errors"
	"interview-cases/case21_30/case24/domain"
	"interview-cases/case21_30/case24/repository/cache/local"
	"interview-cases/case21_30/case24/repository/cache/redis"
	"log/slog"
	"sync/atomic"
	"time"
)

type Cache struct {
	localCache     *local.Cache
	redisCache     *redis.Cache
	cachesStrategy int32 // 缓存策略 0-只有在redis崩溃的时候才会写入本地缓存，1-redis正常的时候也会写入本地缓存
	useLocalCache  int32 // 0-使用redis 1-使用本地缓存
}

func NewCache(localCache *local.Cache, redisCache *redis.Cache, cachesStrategy int32) *Cache {
	c := &Cache{
		localCache:     localCache,
		redisCache:     redisCache,
		cachesStrategy: cachesStrategy,
	}
	go c.redisFailOverLoop()
	return c
}

func (c *Cache) Get(ctx context.Context, id int64) (domain.Order, error) {
	useLocalCache := atomic.LoadInt32(&c.useLocalCache)
	switch useLocalCache {
	case 0:
		return c.redisCache.Get(ctx, id)
	case 1:
		return c.localCache.Get(ctx, id)
	default:
		return domain.Order{}, errors.New("未知状态")
	}
}

func (c *Cache) Set(ctx context.Context, order domain.Order) error {
	switch c.cachesStrategy {
	case 0:
		if atomic.LoadInt32(&c.useLocalCache) == 0 {
			// redis没崩溃就写入redis
			return c.redisCache.Set(ctx, order)
		}
		// 崩溃就写入本地缓存
		return c.localCache.Set(ctx, order)
	case 1:
		err := c.localCache.Set(ctx, order)
		if err != nil {
			return err
		}
		if atomic.LoadInt32(&c.useLocalCache) == 0 {
			return c.redisCache.Set(ctx, order)
		}
	}
	return nil
}

func (c *Cache) Del(ctx context.Context, id int64) error {
	switch c.cachesStrategy {
	case 0:
		if atomic.LoadInt32(&c.useLocalCache) == 0 {
			// redis没崩溃就写入redis
			return c.redisCache.Del(ctx, id)
		}
		// 崩溃就写入本地缓存
		return c.localCache.Del(ctx, id)
	case 1:
		err := c.localCache.Del(ctx, id)
		if err != nil {
			return err
		}
		if atomic.LoadInt32(&c.useLocalCache) == 0 {
			return c.redisCache.Del(ctx, id)
		}
	}
	return nil
}

// redisFailOverLoop
// 检测每隔1s给redis发送ping请求，如果请求失败，隔10ms发送ping请求，
// 如果连续三次不行就故障转移，将请求切换到本地缓存。
// 然后继续每隔一秒给redis发送ping请求连续三次请求成功就切换回redis。
// redisFailOverLoop 负责监控 Redis 的状态并处理故障转移和恢复逻辑
func (c *Cache) redisFailOverLoop() {
	successCount := 0
	for {
		if c.isRedisAvailable() {
			c.handleRecovery(&successCount)
		} else {
			c.handleFailOver()
			successCount = 0
		}
		time.Sleep(1 * time.Second)
	}
}

// 校验 redis 是否崩溃
func (c *Cache) checkRedis() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return c.redisCache.Ping(ctx)
}

// 判断 Redis 是否可用
func (c *Cache) isRedisAvailable() bool {
	err := c.checkRedis()
	if err == nil {
		return true
	}
	// 如果检测到故障，隔 10ms 发送两次 ping 请求进行确认
	for i := 0; i < 2; i++ {
		time.Sleep(10 * time.Millisecond)
		if c.checkRedis() == nil {
			return true
		}
	}
	return false
}

// 处理故障转移逻辑
func (c *Cache) handleFailOver() {
	if ok := atomic.CompareAndSwapInt32(&c.useLocalCache, 0, 1); ok {
		slog.Error("发现redis故障，切换成本地缓存", slog.Any("err", errors.New("redis故障")))
	}
}

// 处理从本地缓存恢复到 Redis 的逻辑
func (c *Cache) handleRecovery(successCount *int) {
	if atomic.LoadInt32(&c.useLocalCache) == 1 {
		*successCount++
		if *successCount >= 3 {
			slog.Info("Redis 连接恢复，切换回 Redis...")
			atomic.StoreInt32(&c.useLocalCache, 0)
			*successCount = 0
		}
	}
}

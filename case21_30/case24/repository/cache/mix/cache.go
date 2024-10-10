package mix

import (
	"context"
	"fmt"
	"interview-cases/case21_30/case24/repository/cache/local"
	"interview-cases/case21_30/case24/repository/cache/redis"
	"sync/atomic"
	"time"
)

type Cache struct {
	localCache    *local.Cache
	redisCache    *redis.Cache
	useLocalCache int32 // 0-使用redis 1-使用本地缓存
}

func (c *Cache) Get(ctx context.Context, key string) (any, error) {
	return nil, nil
}

func (c *Cache) Set(ctx context.Context, key string, val any) error {
	//TODO implement me
	panic("implement me")
}

// redisFailOverHandler
// 检测每隔1s给redis发送ping请求，如果请求失败，隔10ms发送ping请求，
// 如果连续三次不行就故障转移，将请求切换到本地缓存。
// 然后继续每隔一秒给redis发送ping请求连续三次请求成功就切换回redis。
// redisFailOverHandler 负责监控 Redis 的状态并处理故障转移和恢复逻辑
func (c *Cache) redisFailOverHandler() {
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
	if atomic.LoadInt32(&c.useLocalCache) == 0 {
		c.failOver()
	}
}

// 处理从本地缓存恢复到 Redis 的逻辑
func (c *Cache) handleRecovery(successCount *int) {
	if atomic.LoadInt32(&c.useLocalCache) == 1 {
		*successCount++
		if *successCount >= 3 {
			fmt.Println("Redis 连接恢复，切换回 Redis...")
			c.recoverFromFailOver()
			*successCount = 0
		}
	}
}

// 故障转移到本地缓存
func (c *Cache) failOver() {
	atomic.StoreInt32(&c.useLocalCache, 1)
}

// 从故障中恢复，切换回 Redis
func (c *Cache) recoverFromFailOver() {
	atomic.StoreInt32(&c.useLocalCache, 0)
}

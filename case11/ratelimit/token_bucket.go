package ratelimiter

import (
	"sync"
	"time"
)

// TokenBucket 代表一个令牌桶限流器
type TokenBucket struct {
	mu          sync.Mutex
	capacity    int64 // 桶的最大容量
	tokens      int64 // 当前令牌数
	rate        int64 // 每秒生成的令牌数
	lastUpdated time.Time
}

// NewTokenBucket 创建一个新的令牌桶限流器
func NewTokenBucket(capacity, rate int64) *TokenBucket {
	return &TokenBucket{
		capacity:    capacity,
		tokens:      capacity, // 初始化时满桶
		rate:        rate,
		lastUpdated: time.Now(),
	}
}

// Consume 尝试消费指定数量的令牌
func (tb *TokenBucket) Consume(tokens int64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 计算自从上次更新以来应该添加的令牌数量
	tokenToAdd := tb.rate * int64(time.Since(tb.lastUpdated).Seconds())
	if tokenToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokenToAdd)
		tb.lastUpdated = time.Now()
	}

	if tb.tokens < tokens {
		return false // 不足，拒绝请求
	}

	tb.tokens -= tokens
	return true // 允许请求
}

// Tokens 返回还剩余多少令牌
func (tb *TokenBucket) Tokens() int64 {
	return tb.tokens
}

// Add 往令牌桶手动添加令牌 仅用于测试
func (tb *TokenBucket) Add(count int64) {
	tb.tokens += count
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

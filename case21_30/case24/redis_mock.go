package case24

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisMock struct {
	startTime time.Time
	redis.Cmdable
}

func NewRedisMock(client redis.Cmdable) *RedisMock {
	return &RedisMock{
		startTime: time.Now(),
		Cmdable: client,
	}
}

func (r *RedisMock) Ping(ctx context.Context) *redis.StatusCmd {
	// 过了十秒钟模拟崩溃
	now := time.Now()
	if now.Sub(r.startTime) > 10*time.Second && now.Sub(r.startTime) < 15*time.Second {
		cmd := redis.NewStatusCmd(ctx)
		cmd.SetVal("ping不通了")
		return cmd
	} else {
		return r.Cmdable.Ping(ctx)
	}
}

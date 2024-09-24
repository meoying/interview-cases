package case17

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	//go:embed slide_window.lua
	script string
)

type Limiter struct {
	client redis.Cmdable
	window time.Duration
	limit  int
	key    string
}

func NewLimiter(client redis.Cmdable,
	window time.Duration,
	limit int, key string) *Limiter {
	return &Limiter{
		client: client,
		window: window,
		limit:  limit,
		key:    key,
	}
}

func (l Limiter) Allow(ctx context.Context) (bool, error) {
	now := time.Now().UnixMilli()
	val, err := l.client.Eval(ctx, script, []string{l.key},
		l.limit, l.window.Milliseconds(), now).Int()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

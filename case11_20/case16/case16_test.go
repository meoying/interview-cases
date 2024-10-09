package case16

import (
	"context"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"interview-cases/case11_20/case16/cache"
	"testing"
	"time"
)

func TestRedisManager(t *testing.T) {
	// 模拟 Redis 节点 a 和 b
	rdbA, mockA := redismock.NewClientMock()
	rdbB, mockB := redismock.NewClientMock()

	key := "test_key"
	val := "test_value"

	testCases := []struct {
		name string

		ctx context.Context

		before func() *cache.RedisManager
		after  func(context.Context, *cache.RedisManager) (string, error)

		wantRes string
		wantErr error
	}{
		{
			name: "main节点无异常",
			ctx:  context.Background(),
			before: func() *cache.RedisManager {
				mockA.ExpectPing().SetVal("OK")
				mockA.ExpectSet(key, val, 0).SetVal("OK")
				mockA.ExpectGet(key).SetVal(val)

				return cache.NewRedisManager(rdbA, rdbB, 2, 2)
			},
			after: func(ctx context.Context, manager *cache.RedisManager) (string, error) {
				time.Sleep(time.Second)
				err := manager.SetValue(ctx, key, val, 0)
				if err != nil {
					return "", err
				}
				res, err := manager.GetValue(ctx, key)
				return res, err
			},
			wantRes: val,
		},
		{
			name: "main节点异常，切换到backup节点",
			ctx:  context.Background(),
			before: func() *cache.RedisManager {
				mockA.ExpectPing().SetErr(redis.Nil)
				mockA.ExpectPing().SetErr(redis.Nil)

				mockB.ExpectSet(key, val, 0).SetVal("OK")
				mockB.ExpectGet(key).SetVal(val)

				return cache.NewRedisManager(rdbA, rdbB, 2, 2)
			},
			after: func(ctx context.Context, manager *cache.RedisManager) (string, error) {
				time.Sleep(2 * time.Second)

				err := manager.SetValue(ctx, key, val, 0)
				if err != nil {
					return "", err
				}
				res, err := manager.GetValue(ctx, key)
				return res, err
			},
			wantRes: val,
		},
		{
			name: "main节点恢复，切换回backup节点",
			ctx:  context.Background(),
			before: func() *cache.RedisManager {
				mockA.ExpectPing().SetErr(redis.Nil)
				mockA.ExpectPing().SetErr(redis.Nil)
				mockA.ExpectPing().SetVal("OK")
				mockA.ExpectPing().SetVal("OK")

				mockA.ExpectSet(key, val, 0).SetVal("OK")
				mockA.ExpectGet(key).SetVal(val)

				mockB.ExpectSet(key, val, 0).SetVal("OK")
				mockB.ExpectGet(key).SetVal(val)

				return cache.NewRedisManager(rdbA, rdbB, 2, 2)
			},
			after: func(ctx context.Context, manager *cache.RedisManager) (string, error) {
				time.Sleep(4 * time.Second)

				err := manager.SetValue(ctx, key, val, 0)
				if err != nil {
					return "", err
				}
				res, err := manager.GetValue(ctx, key)
				return res, err
			},
			wantRes: val,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rdbManager := tc.before()
			go rdbManager.HeartbeatChecker(tc.ctx)
			res, err := tc.after(tc.ctx, rdbManager)
			assert.Equal(t, tc.wantErr, err)

			if err != nil {
				return
			}

			assert.Equal(t, tc.wantRes, res)
		})
	}
}

package case16

import (
	"context"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"interview-cases/case11_20/case16/cache"
	"testing"
)

func TestCase16(t *testing.T) {
	clientA, mockA := redismock.NewClientMock()
	clientB, mockB := redismock.NewClientMock()

	cKey := "test_key"
	data := "cache"

	testCases := []struct {
		name    string
		ctx     context.Context
		before  func() *cache.Redis
		wantRes string
		wantErr error
	}{
		{
			name: "节点A正常，读写缓存",
			ctx:  context.Background(),
			before: func() *cache.Redis {
				mockA.ExpectPing().SetVal("PONG")
				mockA.ExpectGet(cKey).SetErr(redis.Nil)
				mockA.ExpectPing().SetVal("PONG")
				mockA.ExpectSet(cKey, data, 0).SetVal("OK")
				mockA.ExpectPing().SetVal("PONG")
				mockA.ExpectGet(cKey).SetVal(data)

				r := cache.NewRedis(clientA, clientB)
				return r
			},
			wantRes: data,
		},
		{
			name: "节点A异常，读写缓存",
			before: func() *cache.Redis {
				mockA.ExpectPing().SetErr(redis.Nil)
				mockB.ExpectGet(cKey).SetErr(redis.Nil)
				mockA.ExpectPing().SetErr(redis.Nil)
				mockB.ExpectSet(cKey, data, 0).SetVal("OK")
				mockA.ExpectPing().SetErr(redis.Nil)
				mockB.ExpectGet(cKey).SetVal(data)

				r := cache.NewRedis(clientA, clientB)
				return r
			},
			wantRes: data,
		},
		{
			name: "双节点异常",
			before: func() *cache.Redis {
				mockA.ExpectPing().SetErr(redis.Nil)
				mockB.ExpectGet(cKey).SetErr(redis.Nil)
				mockA.ExpectPing().SetErr(redis.Nil)
				mockB.ExpectSet(cKey, data, 0).SetErr(redis.Nil)

				r := cache.NewRedis(clientA, clientB)
				return r
			},
			wantErr: redis.Nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.before()
			c, err := r.Get(tc.ctx, cKey)

			// 不存在则写入缓存
			if err != nil && err == redis.Nil {
				err := r.Set(tc.ctx, cKey, data)
				assert.Equal(t, tc.wantErr, err)
				if err != nil {
					return
				}
			}

			c, err = r.Get(tc.ctx, cKey)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, c)
		})
	}
}

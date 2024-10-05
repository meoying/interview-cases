package case16

import (
	"context"
	"fmt"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"interview-cases/case11_20/case16/cache"
	"testing"
	"time"
)

func TestRedisManager_Normal(t *testing.T) {
	// 模拟 Redis 节点 a 和 b
	rdbA, mockA := redismock.NewClientMock()
	rdbB, _ := redismock.NewClientMock()

	key := "test_key"
	val := "test_value"

	// 模拟 ping a 节点成功
	mockA.ExpectPing().SetVal("PONG")
	mockA.ExpectSet(key, val, 0).SetVal("OK")
	mockA.ExpectGet(key).SetVal(val)

	ctx := context.Background()
	rm := cache.NewRedisManager(rdbA, rdbB)

	go rm.HeartbeatChecker(ctx)

	time.Sleep(time.Second)

	err := rm.SetValue(ctx, key, val, 0)
	require.NoError(t, err)

	res, err := rm.GetValue(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, val, res)
}

func TestRedisManager_mainNodeErr(t *testing.T) {
	// 模拟 Redis 节点 a 和 b
	rdbA, mockA := redismock.NewClientMock()
	rdbB, mockB := redismock.NewClientMock()

	key := "test_key"
	val := "test_value"

	// 模拟 ping a 节点成功
	mockA.ExpectPing().SetVal("PONG")
	mockB.ExpectSet(key, val, 0).SetVal("OK")
	mockB.ExpectGet(key).SetVal(val)

	ctx := context.Background()
	rm := cache.NewRedisManager(rdbA, rdbB)

	go rm.HeartbeatChecker(ctx)

	time.Sleep(time.Second)

	rm.SimulateFailure()
	err := rm.SetValue(ctx, key, val, 0)
	require.NoError(t, err)

	res, err := rm.GetValue(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, val, res)
}

func TestRedisManager_bothErr(t *testing.T) {
	// 模拟 Redis 节点 a 和 b
	rdbA, mockA := redismock.NewClientMock()
	rdbB, mockB := redismock.NewClientMock()

	key := "test_key"
	val := "test_value"

	// 模拟 ping a 节点成功
	mockA.ExpectPing().SetVal("PONG")
	mockB.ExpectSet(key, val, 0).SetErr(redis.Nil)

	ctx := context.Background()
	rm := cache.NewRedisManager(rdbA, rdbB)

	go rm.HeartbeatChecker(ctx)

	time.Sleep(time.Second)

	rm.SimulateFailure()
	err := rm.SetValue(ctx, key, val, 0)
	assert.Equal(t, fmt.Errorf("键 %s 不存在", key), err)
}

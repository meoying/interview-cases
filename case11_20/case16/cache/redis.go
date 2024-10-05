package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"sync"
	"time"
)

// RedisNode 结构体，代表一个 Redis 节点
type RedisNode struct {
	client   *redis.Client
	isActive bool // 节点是否处于活动状态
}

// RedisManager 结构体，负责管理两个 Redis 节点
type RedisManager struct {
	mainNode      *RedisNode
	backupNode    *RedisNode
	trafficWeight int // a 节点的流量百分比
	lock          sync.Mutex
}

// NewRedisManager 创建 RedisManager 并初始化两个节点
func NewRedisManager(mainClient, backupClient *redis.Client) *RedisManager {
	mainNode := &RedisNode{
		client:   mainClient,
		isActive: true,
	}
	backupNode := &RedisNode{
		client:   backupClient,
		isActive: true,
	}

	return &RedisManager{
		mainNode:      mainNode,
		backupNode:    backupNode,
		trafficWeight: 100, // 初始情况下，100%流量到 a 节点
	}
}

// GetValue 获取 Redis 数据
func (rm *RedisManager) GetValue(ctx context.Context, key string) (string, error) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	selectedNode := rm.getNode()

	val, err := selectedNode.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// 这里预防redis穿透的情况
		return "", fmt.Errorf("键 %s 不存在", key)
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

// SetValue 缓存 Redis 数据
func (rm *RedisManager) SetValue(ctx context.Context, key string, val any, expiration time.Duration) error {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	selectedNode := rm.getNode()

	err := selectedNode.client.Set(ctx, key, val, expiration).Err()
	if err == redis.Nil {
		return fmt.Errorf("键 %s 不存在", key)
	}
	if err != nil {
		return err
	}

	return nil
}

func (rm *RedisManager) getNode() *RedisNode {
	if rm.trafficWeight == 100 || rand.Intn(100) < rm.trafficWeight {
		return rm.mainNode
	} else {
		return rm.backupNode
	}
}

// HeartbeatChecker 定期检测 nodeA 是否恢复
func (rm *RedisManager) HeartbeatChecker(ctx context.Context) {
	for {
		rm.lock.Lock()

		err := rm.pingMainNode(ctx)
		if err == nil {
			if rm.mainNode.isActive == false {
				rm.mainNode.isActive = true
				rm.trafficWeight += 1
				fmt.Println("a 节点已恢复")
			} else if rm.trafficWeight < 100 {
				rm.trafficWeight += 1
				fmt.Printf("将 %d%% 流量切换到 a 节点\n", rm.trafficWeight)
			} else {
				fmt.Println("心跳正常")
			}
		} else {
			fmt.Println("a 节点仍然不可用")
		}

		rm.lock.Unlock()
		time.Sleep(3 * time.Second) // 每隔 1 秒检查一次
	}
}

// pingMainNode 检测 mainNode 的状态
func (rm *RedisManager) pingMainNode(ctx context.Context) error {
	_, err := rm.mainNode.client.Ping(ctx).Result()
	return err
}

// SimulateFailure 模拟 a 节点异常
func (rm *RedisManager) SimulateFailure() {
	rm.lock.Lock()
	rm.mainNode.isActive = false
	rm.trafficWeight = 0 // 切换到 b 节点
	rm.lock.Unlock()
	fmt.Println("a 节点出现异常，流量切换到 b 节点")
}

// SimulateRecovery 模拟 a 节点恢复
func (rm *RedisManager) SimulateRecovery() {
	rm.lock.Lock()
	rm.mainNode.isActive = true
	rm.lock.Unlock()
	fmt.Println("a 节点恢复")
}

package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"sync"
	"time"
)

const (
	MaxTrafficWeight       = 100
	AdjustTrafficWeightCnt = 10

	NoneOfGrayMode      = 0
	Main2BackupGrayMode = 1
	Backup2MainGrayMode = 2
)

var (
	ErrMainNode = errors.New("main节点异常")
)

// RedisNode 结构体，代表一个 Redis 节点
type RedisNode struct {
	client   *redis.Client
	isActive bool // 节点是否处于活动状态
}

// RedisManager 结构体，负责管理两个 Redis 节点
type RedisManager struct {
	mainNode   *RedisNode
	backupNode *RedisNode
	grayMode   int

	mainBreakNum   int
	mainRecoverNum int
	heartbreakCnt  int
	heartbeatCnt   int
	successRespCnt int

	trafficWeight int // 目前节点流量权重

	lock sync.Mutex
}

// NewRedisManager 创建 RedisManager 并初始化两个节点
func NewRedisManager(mainClient, backupClient *redis.Client, mainBreakNum, mainRecoverNum int) *RedisManager {
	mainNode := &RedisNode{
		client:   mainClient,
		isActive: true,
	}
	backupNode := &RedisNode{
		client:   backupClient,
		isActive: true,
	}

	return &RedisManager{
		mainNode:       mainNode,
		backupNode:     backupNode,
		trafficWeight:  MaxTrafficWeight,
		mainBreakNum:   mainBreakNum,
		mainRecoverNum: mainRecoverNum,
	}
}

// GetValue 获取 Redis 数据
func (rm *RedisManager) GetValue(ctx context.Context, key string) (string, error) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	selectedNode, isGrayNode := rm.getNode()
	if selectedNode == nil {
		return "", ErrMainNode
	}

	val, err := selectedNode.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// 这里预防redis穿透的情况
		return "", fmt.Errorf("键 %s 不存在", key)
	}
	if err != nil {
		return "", err
	}

	if isGrayNode {
		rm.adjustWeight()
	}

	return val, nil
}

// SetValue 缓存 Redis 数据
func (rm *RedisManager) SetValue(ctx context.Context, key string, val any, expiration time.Duration) error {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	selectedNode, isGrayNode := rm.getNode()
	if selectedNode == nil {
		return ErrMainNode
	}

	err := selectedNode.client.Set(ctx, key, val, expiration).Err()
	if err == redis.Nil {
		return fmt.Errorf("键 %s 不存在", key)
	}
	if err != nil {
		return err
	}

	if isGrayNode {
		rm.adjustWeight()
	}

	return nil
}

func (rm *RedisManager) getNode() (node *RedisNode, isGrayNode bool) {
	if rm.grayMode == NoneOfGrayMode {
		if rm.mainNode.isActive == true {
			node = rm.mainNode
			return
		} else {
			node = rm.backupNode
			return
		}
	} else if rm.grayMode == Main2BackupGrayMode {
		if rand.Intn(100) < rm.trafficWeight {
			node = rm.backupNode
			isGrayNode = true
			return
		} else {
			return // 拒绝访问
		}
	} else {
		if rand.Intn(100) < rm.trafficWeight {
			node = rm.mainNode
			isGrayNode = true
			return
		} else {
			node = rm.backupNode
			return
		}
	}
}

// adjustWeight 调整灰度节点权重
func (rm *RedisManager) adjustWeight() {
	if rm.grayMode != NoneOfGrayMode {
		rm.successRespCnt += 1
		if rm.successRespCnt%AdjustTrafficWeightCnt == 0 {
			rm.trafficWeight += 1
		}

		if rm.trafficWeight == MaxTrafficWeight {
			rm.grayMode = NoneOfGrayMode
			rm.successRespCnt = 0
		}
	}
}

// HeartbeatChecker 定期检测 nodeA 健康状态
// main切换到backup的灰度需要拒绝main上的所有请求
// backup切换到main的灰度需要
func (rm *RedisManager) HeartbeatChecker(ctx context.Context) {
	for {
		err := rm.pingMainNode(ctx)
		if err == nil {
			if rm.mainNode.isActive == false {
				rm.heartbreakCnt = 0
				rm.heartbeatCnt++
				if rm.heartbeatCnt >= rm.mainRecoverNum {
					rm.mainNode.isActive = true
					rm.heartbeatCnt = 0
					rm.grayMode = Backup2MainGrayMode
					rm.trafficWeight = 1
					fmt.Println("a 节点已恢复，进入灰度模式")
				}
			} else {
				fmt.Println("心跳正常")
			}
		} else {
			if rm.mainNode.isActive == true {
				rm.heartbeatCnt = 0
				rm.heartbreakCnt++
				if rm.heartbreakCnt >= rm.mainBreakNum {
					rm.mainNode.isActive = false
					rm.heartbreakCnt = 0
					rm.grayMode = Main2BackupGrayMode
					rm.trafficWeight = 90 // 为了能快速切换到backup节点
					fmt.Println("main节点不可用，临时切换到backup节点")
				}
			}
		}

		time.Sleep(time.Second) // 每隔 1 秒检查一次
	}
}

// pingMainNode 检测 mainNode 的状态
func (rm *RedisManager) pingMainNode(ctx context.Context) error {
	_, err := rm.mainNode.client.Ping(ctx).Result()
	return err
}

package case32

import (
	"context"
	"sync"
	"time"
)

// NormalAdaptiveStrategy 普通版自适应策略
// 滑动窗口
type NormalAdaptiveStrategy struct {
	mu *sync.RWMutex
	// 重试间隔上限啊
	strategy    Strategy
	slideWindow []req
	// 滑动窗口的阈值
	interval time.Duration
	// 请求允许的失败次数
	failNum int
}

func NewNormalAdaptiveStrategy(strategy Strategy, interval time.Duration, failNum int) *NormalAdaptiveStrategy {
	return &NormalAdaptiveStrategy{
		mu:          &sync.RWMutex{},
		strategy:    strategy,
		slideWindow: make([]req, 0),
		failNum:     failNum,
		interval:    interval,
	}
}

type req struct {
	// 请求的时间戳
	timestamp int64
	//
	success bool
}

func (n *NormalAdaptiveStrategy) Next(ctx context.Context, err error) (time.Duration, bool) {
	if err == nil {
		return n.success(ctx, err)
	} else {
		return n.fail(ctx, err)
	}
}

func (n *NormalAdaptiveStrategy) success(ctx context.Context, err error) (time.Duration, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.slideWindow = append(n.slideWindow, req{
		timestamp: time.Now().UnixMilli(),
		success:   true,
	})
	return n.strategy.Next(ctx, err)
}

func (n *NormalAdaptiveStrategy) fail(ctx context.Context, err error) (time.Duration, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	now := time.Now().UnixMilli()
	threshold := now - n.interval.Milliseconds()
	validIdx := 0
	// 剔除超过时间的请求
	for i := 0; i < len(n.slideWindow); i++ {
		if n.slideWindow[i].timestamp >= threshold {
			n.slideWindow[validIdx] = n.slideWindow[i]
			validIdx++
		}
	}
	n.slideWindow = n.slideWindow[:validIdx]
	// 统计失败请求数
	failCount := 0
	for _, r := range n.slideWindow {
		if !r.success {
			failCount++
		}
	}
	// 添加失败请求
	n.slideWindow = append(n.slideWindow, req{
		timestamp: now,
		success:   false,
	})

	if failCount >= n.failNum {
		// 如果超过阈值就不重试了
		return 0, false
	}
	return n.strategy.Next(ctx, err)
}

// getCount 测试用
func (n *NormalAdaptiveStrategy) getCount() (int, int) {
	successCount := 0
	failCount := 0
	for _, r := range n.slideWindow {
		if r.success {
			successCount++
		} else {
			failCount++
		}
	}
	return successCount, failCount
}

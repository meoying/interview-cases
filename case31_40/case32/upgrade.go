package case32

import (
	"context"
	"math/bits"
	"sync/atomic"
	"time"
)

// UpgradeAdaptiveStrategy 进阶版自适应策略
// 使用ring buffer优化
type UpgradeAdaptiveStrategy struct {
	s          Strategy
	threshold  int // 比特数量
	ringBuffer []uint64
	reqCount   uint64
}

func NewUpgradeAdaptiveStrategy(s Strategy, threshold int) *UpgradeAdaptiveStrategy {
	return &UpgradeAdaptiveStrategy{
		s:         s,
		threshold: threshold,
		// 初始化16个uint64 表示1024个请求
		ringBuffer: make([]uint64, 16),
	}
}

func (u *UpgradeAdaptiveStrategy) Next(ctx context.Context, err error) (time.Duration, bool) {
	if err == nil {
		u.markSuccess()
		return u.s.Next(ctx, err)
	} else {
		failCount := u.getFailed()
		u.markFail()
		if failCount >= u.threshold {
			return 0, false
		}
		return u.s.Next(ctx, err)
	}
}

func (u *UpgradeAdaptiveStrategy) getFailed() int {
	var failCount int
	for i := 0; i < len(u.ringBuffer); i++ {
		v := atomic.LoadUint64(&u.ringBuffer[i])
		failCount += bits.OnesCount64(v)
	}
	return failCount
}

func (u *UpgradeAdaptiveStrategy) markSuccess() {
	count := atomic.AddUint64(&u.reqCount, 1)
	count = count % 1024
	// 确定在ringBuffer中的索引
	index := count / 64
	// 确定在uint64中的位
	bitPosition := uint(count % 64)

	// 使用原子操作将该位设置为0
	for {
		old := atomic.LoadUint64(&u.ringBuffer[index])
		ne := old &^ (uint64(1) << bitPosition) // 使用按位清除操作
		if atomic.CompareAndSwapUint64(&u.ringBuffer[index], old, ne) {
			break
		}
	}
}

func (u *UpgradeAdaptiveStrategy) markFail() {
	count := atomic.AddUint64(&u.reqCount, 1)
	count = count % 1024
	// 确定在ringBuffer中的索引
	index := count / 64
	// 确定在uint64中的位
	bitPosition := uint(count % 64)

	for {
		old := atomic.LoadUint64(&u.ringBuffer[index])
		ne := old | (uint64(1) << bitPosition) // 使用按位或操作来设置1
		if atomic.CompareAndSwapUint64(&u.ringBuffer[index], old, ne) {
			break
		}
	}
}

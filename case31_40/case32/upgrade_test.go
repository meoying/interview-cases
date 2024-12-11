package case32

import (
	"context"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"testing"
)

// 测试场景
// 阈值是50
// 2000个请求 有1500个成功的 有500个失败的 最后统计500个失败的有50个可以执行 有450个不能执行 1500成功的都能执行
func TestUpgradeAdaptiveStrategy_Next_Concurrent(t *testing.T) {
	// 创建一个基础策略
	baseStrategy := &MockStrategy{}

	// 创建升级版自适应策略，设置阈值为50
	strategy := NewUpgradeAdaptiveStrategy(baseStrategy, 50)

	var wg sync.WaitGroup
	var successCount, errCount int64
	mockErr := errors.New("mock error")

	// 并发执行2000个请求
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// 前1500个请求成功，后500个失败
			var err error
			if index >= 1500 {
				err = mockErr
			}

			_, allowed := strategy.Next(context.Background(), err)

			if err != nil {
				// 失败请求的统计
				if allowed {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&errCount, 1)
				}
			} else {
				// 成功请求必须被允许
				if !allowed {
					t.Errorf("预期成功请求应该被允许执行，但第%d个请求被拒绝", index+1)
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 验证结果：期望大约50个失败请求可以执行，450个被拒绝
	// 由于是环形缓冲区和并发执行，可能会有一些误差，这里使用一个合理的范围进行判断
	finalSuccessCount := int(atomic.LoadInt64(&successCount))
	finalErrCount := int(atomic.LoadInt64(&errCount))
	if finalSuccessCount < 45 || finalSuccessCount > 55 {
		t.Errorf("期望大约50个失败请求被允许执行，实际允许执行的失败请求数量为: %d", finalSuccessCount)
	}

	if finalErrCount < 445 || finalErrCount > 455 {
		t.Errorf("期望大约450个失败请求被拒绝执行，实际被拒绝的失败请求数量为: %d", finalErrCount)
	}
}

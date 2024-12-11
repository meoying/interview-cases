package case32

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockStrategy struct {
}

func (m MockStrategy) Next(ctx context.Context, err error) (time.Duration, bool) {
	return 1 * time.Second, true
}

// 测试场景 间隔1s 最多20个失败请求
// 10个成功的请求 20个失败请求 这20个请求调用next是可以继续的。第21个调用next不可以继续。 睡到600ms
// 然后再发 10个成功的请求 10个失败的请求 断言成功的请求next可以继续，失败的请求next是不能继续
// 再睡500ms 失败请求断言可以继续 断言有 slicewindow有10个成功的请求 11个失败请求
func TestNormalAdaptiveStrategy_Next_FailuresWithinThreshold(t *testing.T) {
	// 初始化策略
	s := NewNormalAdaptiveStrategy(&MockStrategy{}, time.Second, 20)
	var wg sync.WaitGroup
	// 并发发起10个成功请求
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, ok := s.Next(context.Background(), nil)
			assert.True(t, ok, "成功请求应该可以继续")
		}()
	}

	// 并发发起20个失败请求
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, ok := s.Next(context.Background(), fmt.Errorf("mock error"))
			assert.True(t, ok, "前20个失败请求应该可以继续")
		}()
	}

	wg.Wait()

	// 发起第21个失败请求
	_, ok := s.Next(context.Background(), fmt.Errorf("mock error"))
	assert.False(t, ok, "第21个失败请求不应该继续")

	// 等待600ms
	time.Sleep(600 * time.Millisecond)
	// 并发再发10个成功请求
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, ok := s.Next(context.Background(), nil)
			assert.True(t, ok, "成功请求应该可以继续")
		}()
	}

	// 并发再发10个失败请求
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, ok := s.Next(context.Background(), fmt.Errorf("mock error"))
			assert.False(t, ok, "失败请求不应该继续")
		}()
	}

	wg.Wait()

	// 再等待500ms
	time.Sleep(500 * time.Millisecond)

	// 发起一个失败请求
	_, ok = s.Next(context.Background(), fmt.Errorf("mock error"))
	assert.True(t, ok, "失败请求应该可以继续")

	// 验证滑动窗口中有10个成功请求和11个失败请求
	successCount, failCount := s.getCount()
	assert.Equal(t, 10, successCount, "滑动窗口中应该有10个成功请求")
	assert.Equal(t, 11, failCount, "滑动窗口中应该有11个失败请求")
}

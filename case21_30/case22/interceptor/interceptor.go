package interceptor

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"interview-cases/case21_30/case22/monitor"
	"log/slog"
	"sync/atomic"
	"time"
)

type MemoryLimiter struct {
	// 状态 0-正常 1-限流状态
	state int32
	// 限流令牌
	limitCh chan struct{}
	// 获取监控数据的抽象
	mon monitor.Monitor
	// 间隔多久获取监控数据
	interval time.Duration
}

func NewMemoryLimiter(mon monitor.Monitor, interval time.Duration) *MemoryLimiter {
	m := &MemoryLimiter{
		mon: mon,
		// 可以调节这个缓存大小，改变限流的频率
		limitCh:  make(chan struct{}, 1),
		interval: interval,
	}
	go m.monitor()
	return m
}

func (m *MemoryLimiter) monitor() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		usage, err := m.mon.GetMemoryUsage(ctx)
		cancel()
		if err != nil {
			// 记录一下日志
			slog.Error("获取监控信息失败", slog.Any("err", err))
			continue
		}
		// 超内存了
		if usage >= 80 {
			atomic.StoreInt32(&m.state, 1)
		} else {
			atomic.StoreInt32(&m.state, 0)
		}
		time.Sleep(m.interval)
	}

}

func (m *MemoryLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ticker := time.NewTicker(100 * time.Millisecond)
		for atomic.LoadInt32(&m.state) == 1 {
			// 当前处于限流状态
			select {
			case <-ticker.C:
				// 每100ms查看一下当前状态是否为限流状态
			case m.limitCh <- struct{}{}:
				// 一个个来
				resp, err = handler(ctx, req)
				<-m.limitCh
				return
			case <-ctx.Done():
				return nil, errors.New("等待超时退出")
			}
		}
		resp, err = handler(ctx, req)
		return
	}
}

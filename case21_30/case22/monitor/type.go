package monitor

import "context"

// 用于监控内存的接口
type Monitor interface {
	GetMemoryUsage(ctx context.Context) (float64, error)
}

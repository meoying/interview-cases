package monitor

import (
	"context"
	"time"
)

type MockMon struct {
	startTime int64
}

func NewMockMonitor() *MockMon {
	return &MockMon{
		startTime: time.Now().UnixMilli(),
	}
}

func (m *MockMon) GetMemoryUsage(ctx context.Context) (float64, error) {
	nowTime := time.Now().UnixMilli()
	// 2秒后超过80%
	intervalTime := nowTime - m.startTime
	if intervalTime >= 2000 && intervalTime <= 5000 {
		return 0.95, nil
	}
	return 0.5, nil
}

package monitor

import (
	"context"
	"github.com/shirou/gopsutil/mem"
)

// PhysicalMonitor 从物理机直接获取监控信息
type PhysicalMonitor struct {
}

func NewPhysicalMonitor() *PhysicalMonitor {
	return &PhysicalMonitor{}
}

func (p *PhysicalMonitor) GetMemoryUsage(ctx context.Context) (float64, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return v.UsedPercent, nil
}

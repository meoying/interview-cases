package monitor

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPhysicalMonitor_GetMemoryUsage(t *testing.T) {
	monitor := NewPhysicalMonitor()
	usage, err := monitor.GetMemoryUsage(context.Background())
	require.NoError(t, err)
	fmt.Printf("当前物理机的内存使用率为 %.2f%%\n", usage)
}

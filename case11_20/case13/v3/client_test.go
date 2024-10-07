package v3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client, "新创建的客户端不应为nil")
}

func TestAddNode(t *testing.T) {
	client := NewClient()
	url := "http://example.com"
	weight := 50

	client.AddNode(url, weight)

	w, exists := client.GetWeight(url)
	assert.True(t, exists, "应该能找到添加的节点")
	assert.Equal(t, weight, w, "节点的权重应该与设置的一致")
}

func TestAdjustWeight(t *testing.T) {
	client := NewClient()
	url := "http://example.com"
	initialWeight := 50

	testCases := []struct {
		name           string
		err            error
		expectedWeight int
	}{
		{"网络异常", ErrNetworkFailure, 0},
		{"服务熔断", ErrCircuitBreaker, 0},
		{"请求超时", ErrTimeout, 49},
		{"服务限流", ErrThrottling, 25},
		{"无错误", nil, 50},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.AddNode(url, initialWeight)
			client.AdjustWeight(url, tc.err)
			weight, exists := client.GetWeight(url)
			assert.True(t, exists, "节点应该存在")
			assert.Equal(t, tc.expectedWeight, weight, "权重调整后应该符合预期")
		})
	}
}

func TestWeightBounds(t *testing.T) {
	client := NewClient()
	url := "http://example.com"

	// 测试下限
	t.Run("权重下限", func(t *testing.T) {
		client.AddNode(url, MinWeight)
		client.AdjustWeight(url, ErrTimeout)
		weight, _ := client.GetWeight(url)
		assert.Equal(t, MinWeight, weight, "权重不应低于最小值")
	})

	// 测试上限
	t.Run("权重上限", func(t *testing.T) {
		client.AddNode(url, MaxWeight)
		client.AdjustWeight(url, nil)
		weight, _ := client.GetWeight(url)
		assert.Equal(t, MaxWeight, weight, "权重不应超过最大值")
	})
}

package v4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeightedRoundRobinLoadBalancer(t *testing.T) {
	testCases := []struct {
		name          string
		nodes         []*Node
		expectedOrder []string
		expectedError error
	}{
		{
			name: "正常权重分配",
			nodes: []*Node{
				{URL: "node1", Weight: 2},
				{URL: "node2", Weight: 1},
				{URL: "node3", Weight: 1},
			},
			expectedOrder: []string{"node1", "node2", "node3", "node1"},
			expectedError: nil,
		},
		{
			name: "包含零权重节点",
			nodes: []*Node{
				{URL: "node1", Weight: 2},
				{URL: "node2", Weight: 0},
				{URL: "node3", Weight: 1},
			},
			expectedOrder: []string{"node1", "node3", "node1"},
			expectedError: nil,
		},
		{
			name: "所有节点权重为零",
			nodes: []*Node{
				{URL: "node1", Weight: 0},
				{URL: "node2", Weight: 0},
			},
			expectedOrder: []string{},
			expectedError: ErrNoAvailableNodes,
		},
		{
			name:          "空节点列表",
			nodes:         []*Node{},
			expectedOrder: []string{},
			expectedError: ErrNoAvailableNodes,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lb := &WeightedRoundRobinLoadBalancer{}

			var selectedNodes []string
			for i := 0; i < len(tc.expectedOrder); i++ {
				node, err := lb.Select(tc.nodes)

				if tc.expectedError != nil {
					assert.Equal(t, tc.expectedError, err)
					continue
				}

				assert.NoError(t, err)
				assert.NotNil(t, node)
				selectedNodes = append(selectedNodes, node.URL)
			}

			if tc.expectedError == nil {
				assert.Equal(t, tc.expectedOrder, selectedNodes)
			} else {
				assert.Empty(t, selectedNodes)
			}
		})
	}
}

func TestWeightedRoundRobinLoadBalancerCycleCompletion(t *testing.T) {
	nodes := []*Node{
		{URL: "node1", Weight: 2},
		{URL: "node2", Weight: 1},
	}

	lb := &WeightedRoundRobinLoadBalancer{}

	expectedOrder := []string{"node1", "node2", "node1", "node1", "node2", "node1"}
	for i := 0; i < len(expectedOrder); i++ {
		node, err := lb.Select(nodes)
		assert.NoError(t, err)
		assert.Equal(t, expectedOrder[i], node.URL)
	}
}

func TestWeightedRoundRobinLoadBalancerWeightChange(t *testing.T) {
	nodes := []*Node{
		{URL: "node1", Weight: 2},
		{URL: "node2", Weight: 1},
	}

	lb := &WeightedRoundRobinLoadBalancer{}

	// 第一轮选择
	node, _ := lb.Select(nodes)
	assert.Equal(t, "node1", node.URL)

	// 改变节点权重
	nodes[0].Weight = 1
	nodes[1].Weight = 2

	// 重置负载均衡器以应用新的权重
	lb = &WeightedRoundRobinLoadBalancer{}

	expectedOrder := []string{"node2", "node1", "node2", "node2", "node1", "node2"}
	for i := 0; i < len(expectedOrder); i++ {
		node, err := lb.Select(nodes)
		assert.NoError(t, err)
		assert.Equal(t, expectedOrder[i], node.URL)
	}
}

package case27

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestSuite struct {
	suite.Suite
}

func (s *TestSuite) TestLoadBalancer() {
	readLoadBalancer := &WeightedRoundRobinLoadBalancer{}
	writeLoadBalancer := &WeightedRoundRobinLoadBalancer{}
	client := NewRWWeightClient(readLoadBalancer, writeLoadBalancer)
	// 添加写节点
	err := client.AddWriteNode(&ServiceNode{
		URL:    "w1.com",
		Weight: 25,
	})
	require.NoError(s.T(), err)
	err = client.AddWriteNode(&ServiceNode{
		URL:    "w2.com",
		Weight: 20,
	})
	require.NoError(s.T(), err)
	err = client.AddWriteNode(&ServiceNode{
		URL:    "w3.com",
		Weight: 40,
	})
	require.NoError(s.T(), err)
	// 添加读节点
	err = client.AddReadNode(&ServiceNode{
		URL:    "r1.com",
		Weight: 20,
	})
	require.NoError(s.T(), err)
	err = client.AddReadNode(&ServiceNode{
		URL:    "r2.com",
		Weight: 40,
	})
	require.NoError(s.T(), err)
	err = client.AddReadNode(&ServiceNode{
		URL:    "r3.com",
		Weight: 80,
	})
	require.NoError(s.T(), err)
	// 获取读节点
	actualNodes := make([]*ServiceNode, 0, 10)
	for i := 0; i < 10; i++ {
		node, err := client.Get(context.Background())
		require.NoError(s.T(), err)
		actualNodes = append(actualNodes, node)
	}
	assert.Equal(s.T(), []*ServiceNode{
		{
			URL:    "r3.com",
			Weight: 80,
		},
		{
			URL:    "r2.com",
			Weight: 40,
		},
		{
			URL:    "r3.com",
			Weight: 80,
		},
		{
			URL:    "r1.com",
			Weight: 20,
		},
		{
			URL:    "r3.com",
			Weight: 80,
		},
		{
			URL:    "r2.com",
			Weight: 40,
		},
		{
			URL:    "r3.com",
			Weight: 80,
		},
		{
			URL:    "r3.com",
			Weight: 80,
		},
		{
			URL:    "r2.com",
			Weight: 40,
		},
		{
			URL:    "r3.com",
			Weight: 80,
		},
	}, actualNodes)
	// 获取写节点
	actualNodes = make([]*ServiceNode, 0, 10)
	for i := 0; i < 10; i++ {
		//ctx 携带requestType标记位
		node, err := client.Get(s.withNodeType(context.Background()))
		require.NoError(s.T(), err)
		actualNodes = append(actualNodes, node)
	}
	assert.Equal(s.T(), []*ServiceNode{
		{
			URL:    "w3.com",
			Weight: 40,
		},
		{
			URL:    "w1.com",
			Weight: 25,
		},
		{
			URL:    "w2.com",
			Weight: 20,
		},
		{
			URL:    "w3.com",
			Weight: 40,
		},
		{
			URL:    "w1.com",
			Weight: 25,
		},
		{
			URL:    "w3.com",
			Weight: 40,
		},
		{
			URL:    "w2.com",
			Weight: 20,
		},
		{
			URL:    "w3.com",
			Weight: 40,
		},
		{
			URL:    "w1.com",
			Weight: 25,
		},
		{
			URL:    "w3.com",
			Weight: 40,
		},
	}, actualNodes)
}

func (s *TestSuite) withNodeType(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, RequestType, 1)
	return ctx
}

// 测试请求
func TestDelayMsg(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

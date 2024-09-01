package v4

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// TestIntegrationSuite 运行集成测试套件
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// IntegrationTestSuite 定义集成测试套件
type IntegrationTestSuite struct {
	suite.Suite
	client *Client
	nodes  []*httptest.Server
}

// SetupSuite 在所有测试开始前运行，用于设置测试环境
func (s *IntegrationTestSuite) SetupSuite() {
	// 创建模拟服务器
	s.nodes = make([]*httptest.Server, 3)
	for i := 0; i < 3; i++ {
		s.nodes[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	}

}

func (s *IntegrationTestSuite) SetupTest() {
	// 创建客户端
	lb := &WeightedRoundRobinLoadBalancer{} // 假设我们有一个轮询负载均衡器
	client, err := NewClient(1, 100, 10, lb, 5*time.Second)
	s.NoError(err)
	s.client = client

	// 添加节点
	for _, node := range s.nodes {
		s.client.AddNode(node.URL)
	}
}

// TearDownSuite 在所有测试结束后运行，用于清理测试环境
func (s *IntegrationTestSuite) TearDownSuite() {
	for _, node := range s.nodes {
		node.Close()
	}
	s.client.Close()
}

// TestBasicFunctionality 测试基本功能
func (s *IntegrationTestSuite) TestBasicFunctionality() {
	s.Run("NormalWeightDistribution", func() {
		counts := make(map[string]int)
		for i := 0; i < 100; i++ {
			node, err := s.client.GetNode()
			s.Require().NoError(err)
			counts[node.URL]++
		}
		// 验证所有节点都被访问到
		for _, node := range s.nodes {
			s.Assert().Greater(counts[node.URL], 0)
		}
	})

	s.Run("RoundRobinMechanism", func() {
		previousNode, err := s.client.GetNode()
		s.Require().NoError(err)
		for i := 0; i < 10; i++ {
			node, err := s.client.GetNode()
			s.Require().NoError(err)
			s.Assert().NotEqual(previousNode.URL, node.URL)
			previousNode = node
		}
	})
}

// TestNodeStatusManagement 测试节点状态管理
func (s *IntegrationTestSuite) TestNodeStatusManagement() {
	s.Run("HealthyNodeHandling", func() {
		node, err := s.client.GetNode()
		s.Require().NoError(err)
		s.Assert().Equal(StatusHealthy, node.Status)
	})

	s.Run("ProbationNodeHandling", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrTimeout)
		// 验证节点进入观察状态
		s.Assert().Equal(StatusProbation, s.client.findNode(node.URL).Status)
	})

	s.Run("UnhealthyNodeHandling", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrNetworkFailure)
		// 验证节点不健康且不再被选中
		s.Assert().Equal(StatusUnhealthy, s.client.findNode(node.URL).Status)
		for i := 0; i < 10; i++ {
			newNode, _ := s.client.GetNode()
			s.Assert().NotEqual(node.URL, newNode.URL)
		}
	})
}

// TestDynamicWeightAdjustment 测试动态权重调整
func (s *IntegrationTestSuite) TestDynamicWeightAdjustment() {
	s.Run("IncreaseWeightOnSuccess", func() {
		node, _ := s.client.GetNode()
		initialWeight := node.Weight
		s.client.UpdateNodeStatus(node.URL, nil) // 成功请求
		s.Assert().Greater(s.client.findNode(node.URL).Weight, initialWeight)
	})

	s.Run("DecreaseWeightOnTimeout", func() {
		node, _ := s.client.GetNode()
		initialWeight := node.Weight
		s.client.UpdateNodeStatus(node.URL, ErrTimeout)
		s.Assert().Less(s.client.findNode(node.URL).Weight, initialWeight)
	})

	s.Run("HalveWeightOnThrottling", func() {
		node, _ := s.client.GetNode()
		initialWeight := node.Weight
		s.client.UpdateNodeStatus(node.URL, ErrThrottling)
		s.Assert().Equal(initialWeight/2, s.client.findNode(node.URL).Weight)
	})

	s.Run("WeightBoundaries", func() {
		node, _ := s.client.GetNode()
		for i := 0; i < 200; i++ {
			s.client.UpdateNodeStatus(node.URL, nil) // 持续成功
		}
		s.Assert().LessOrEqual(s.client.findNode(node.URL).Weight, s.client.maxWeight)

		for i := 0; i < 200; i++ {
			s.client.UpdateNodeStatus(node.URL, ErrThrottling) // 持续限流
		}
		s.Assert().GreaterOrEqual(s.client.findNode(node.URL).Weight, s.client.minWeight)
	})
}

// TestNodeStateTransition 测试节点状态转换
func (s *IntegrationTestSuite) TestNodeStateTransition() {
	s.Run("HealthyToProbation", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrTimeout)
		s.Assert().Equal(StatusProbation, s.client.findNode(node.URL).Status)
	})

	s.Run("ProbationToHealthy", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrTimeout) // 先变为观察状态
		s.client.UpdateNodeStatus(node.URL, nil)        // 成功请求
		s.Assert().Equal(StatusHealthy, s.client.findNode(node.URL).Status)
	})

	s.Run("HealthyToUnhealthy", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrNetworkFailure)
		s.Assert().Equal(StatusUnhealthy, s.client.findNode(node.URL).Status)
	})

	s.Run("UnhealthyToProbation", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrNetworkFailure) // 先变为不健康
		time.Sleep(s.client.recoveryInterval + time.Second)    // 等待恢复间隔
		s.client.tryRecoverNodes()                             // 手动触发恢复尝试
		s.Assert().Equal(StatusProbation, s.client.findNode(node.URL).Status)
	})
}

// TestErrorHandling 测试错误处理
func (s *IntegrationTestSuite) TestErrorHandling() {
	s.Run("NetworkFailureHandling", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrNetworkFailure)
		s.Assert().Equal(StatusUnhealthy, s.client.findNode(node.URL).Status)
	})

	s.Run("TimeoutHandling", func() {
		node, _ := s.client.GetNode()
		initialWeight := node.Weight
		s.client.UpdateNodeStatus(node.URL, ErrTimeout)
		s.Assert().Equal(StatusProbation, s.client.findNode(node.URL).Status)
		s.Assert().Less(s.client.findNode(node.URL).Weight, initialWeight)
	})

	s.Run("ThrottlingHandling", func() {
		node, _ := s.client.GetNode()
		initialWeight := node.Weight
		s.client.UpdateNodeStatus(node.URL, ErrThrottling)
		s.Assert().Equal(StatusProbation, s.client.findNode(node.URL).Status)
		s.Assert().Equal(initialWeight/2, s.client.findNode(node.URL).Weight)
	})

	s.Run("CircuitBreakerHandling", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrCircuitBreaker)
		s.Assert().Equal(StatusUnhealthy, s.client.findNode(node.URL).Status)
	})
}

// TestNodeRecoveryMechanism 测试节点恢复机制
func (s *IntegrationTestSuite) TestNodeRecoveryMechanism() {
	s.Run("AutomaticRecoveryAttempt", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrNetworkFailure)
		time.Sleep(s.client.recoveryInterval + time.Second)
		s.client.tryRecoverNodes()
		s.Assert().Equal(StatusProbation, s.client.findNode(node.URL).Status)
	})

	s.Run("WeightManagementDuringRecovery", func() {
		node, _ := s.client.GetNode()
		s.client.UpdateNodeStatus(node.URL, ErrNetworkFailure)
		time.Sleep(s.client.recoveryInterval + time.Second)
		s.client.tryRecoverNodes()
		s.Assert().Equal(s.client.minWeight, s.client.findNode(node.URL).Weight)
	})
}

// TestConcurrencySafety 测试并发安全性
func (s *IntegrationTestSuite) TestConcurrencySafety() {
	s.Run("ConcurrentRequests", func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := s.client.GetNode()
				s.Assert().NoError(err)
			}()
		}
		wg.Wait()
	})

	s.Run("ConcurrentStatusUpdates", func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				node, _ := s.client.GetNode()
				if i%2 == 0 {
					s.client.UpdateNodeStatus(node.URL, nil)
				} else {
					s.client.UpdateNodeStatus(node.URL, ErrTimeout)
				}
			}(i)
		}
		wg.Wait()
	})
}

// TestEdgeCases 测试边界条件
func (s *IntegrationTestSuite) TestEdgeCases() {
	s.Run("AllNodesUnavailable", func() {
		for _, node := range s.nodes {
			s.client.UpdateNodeStatus(node.URL, ErrNetworkFailure)
		}
		_, err := s.client.GetNode()
		s.Assert().Equal(ErrNoAvailableNodes, err)
		// 恢复节点状态
		for _, node := range s.nodes {
			s.client.UpdateNodeStatus(node.URL, nil)
		}
	})

	s.Run("SingleNodeScenario", func() {
		// 临时移除其他节点
		backupNodes := s.client.healthyNodes[1:]
		s.client.healthyNodes = s.client.healthyNodes[:1]
		defer func() {
			s.client.healthyNodes = append(s.client.healthyNodes, backupNodes...)
		}()

		for i := 0; i < 10; i++ {
			node, err := s.client.GetNode()
			s.Assert().NoError(err)
			s.Assert().Equal(s.client.healthyNodes[0].URL, node.URL)
		}
	})

	s.Run("MaximumNodeCount", func() {
		for i := 0; i < 1000; i++ {
			s.client.AddNode(fmt.Sprintf("http://test%d.com", i))
		}
		_, err := s.client.GetNode()
		s.Assert().NoError(err)
	})
}

// TestConfigurationValidity 测试配置有效性
func (s *IntegrationTestSuite) TestConfigurationValidity() {

	s.Run("InitializationParameterValidation", func() {
		_, err := NewClient(10, 5, 7, &WeightedRoundRobinLoadBalancer{}, 5*time.Second)
		s.Assert().Error(err)

		_, err = NewClient(1, 100, 50, &WeightedRoundRobinLoadBalancer{}, 5*time.Second)
		s.Assert().NoError(err)
	})

	s.Run("RecoveryIntervalSetting", func() {
		client, _ := NewClient(1, 100, 10, &WeightedRoundRobinLoadBalancer{}, 1*time.Second)
		url := "http://example.com"
		client.AddNode(url)
		node, _ := client.GetNode()
		client.UpdateNodeStatus(node.URL, ErrNetworkFailure)
		time.Sleep(2 * time.Second)
		client.tryRecoverNodes()
		s.Assert().Equal(StatusProbation, client.findNode(node.URL).Status)

		client, _ = NewClient(1, 100, 10, &WeightedRoundRobinLoadBalancer{}, 10*time.Second)
		client.AddNode(url)
		node, _ = client.GetNode()
		client.UpdateNodeStatus(node.URL, ErrNetworkFailure)
		time.Sleep(2 * time.Second)
		client.tryRecoverNodes()
		s.Assert().Equal(StatusUnhealthy, client.findNode(node.URL).Status)
	})
}

/*
// TestLongTermStability 测试长期稳定性
func (s *IntegrationTestSuite) TestLongTermStability() {
	s.Run("ContinuousOperation", func() {
		start := time.Now()
		for time.Since(start) < 5*time.Minute {
			node, err := s.client.GetNode()
			s.Assert().NoError(err)
			if time.Now().UnixNano()%2 == 0 {
				s.client.UpdateNodeStatus(node.URL, nil)
			} else {
				s.client.UpdateNodeStatus(node.URL, ErrTimeout)
			}
			time.Sleep(10 * time.Millisecond)
		}
	})

	s.Run("FrequentStateChanges", func() {
		for i := 0; i < 1000; i++ {
			node, _ := s.client.GetNode()
			if i%3 == 0 {
				s.client.UpdateNodeStatus(node.URL, nil)
			} else if i%3 == 1 {
				s.client.UpdateNodeStatus(node.URL, ErrTimeout)
			} else {
				s.client.UpdateNodeStatus(node.URL, ErrNetworkFailure)
			}
		}
		// 验证系统仍然可以正常工作
		_, err := s.client.GetNode()
		s.Assert().NoError(err)
	})
}
*/

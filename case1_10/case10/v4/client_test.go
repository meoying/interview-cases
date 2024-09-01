package v4

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLoadBalancer 是一个模拟的负载均衡器
type MockLoadBalancer struct {
	mock.Mock
}

func (m *MockLoadBalancer) Select(nodes []*Node) (*Node, error) {
	args := m.Called(nodes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Node), args.Error(1)
}

// TestNewClient 测试客户端创建是否正确
func TestNewClient(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, err := NewClient(1, 10, 5, lb, time.Second)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 1, client.minWeight)
	assert.Equal(t, 10, client.maxWeight)
	assert.Equal(t, 5, client.defaultWeight)
	assert.Equal(t, time.Second, client.recoveryInterval)
}

// TestNewClientInvalidWeights 测试使用无效权重创建客户端时是否返回错误
func TestNewClientInvalidWeights(t *testing.T) {
	lb := new(MockLoadBalancer)
	_, err := NewClient(10, 5, 7, lb, time.Second)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的权重配置")
}

// TestAddNode 测试添加节点功能
func TestAddNode(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)

	client.AddNode("http://example.com")

	assert.Len(t, client.healthyNodes, 1)
	assert.Equal(t, "http://example.com", client.healthyNodes[0].URL)
	assert.Equal(t, 5, client.healthyNodes[0].Weight)
	assert.Equal(t, StatusHealthy, client.healthyNodes[0].Status)
}

// TestGetNode 测试获取节点功能
func TestGetNode(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)
	client.AddNode("http://example.com")

	expectedNode := client.healthyNodes[0]
	lb.On("Select", mock.Anything).Return(expectedNode, nil)

	node, err := client.GetNode()

	assert.NoError(t, err)
	assert.Equal(t, expectedNode, node)
	lb.AssertExpectations(t)
}

// TestGetNodeNoAvailableNodes 测试当没有可用节点时的行为
func TestGetNodeNoAvailableNodes(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)

	node, err := client.GetNode()

	assert.Error(t, err)
	assert.Nil(t, node)
	assert.Equal(t, ErrNoAvailableNodes, err)
}

// TestUpdateNodeStatus_Success 测试成功更新节点状态
func TestUpdateNodeStatus_Success(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)
	client.AddNode("http://example.com")

	client.UpdateNodeStatus("http://example.com", nil)

	assert.Len(t, client.healthyNodes, 1)
	assert.Equal(t, StatusHealthy, client.healthyNodes[0].Status)
	assert.Equal(t, 6, client.healthyNodes[0].Weight)
}

// TestUpdateNodeStatus_NetworkFailure 测试网络故障时的节点状态更新
func TestUpdateNodeStatus_NetworkFailure(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)
	client.AddNode("http://example.com")

	client.UpdateNodeStatus("http://example.com", ErrNetworkFailure)

	assert.Len(t, client.unhealthyNodes, 1)
	assert.Equal(t, StatusUnhealthy, client.unhealthyNodes[0].Status)
	assert.Equal(t, 0, client.unhealthyNodes[0].Weight)
}

// TestUpdateNodeStatus_CircuitBreaker 测试熔断器开启时的节点状态更新
func TestUpdateNodeStatus_CircuitBreaker(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)
	client.AddNode("http://example.com")

	client.UpdateNodeStatus("http://example.com", ErrCircuitBreaker)

	assert.Len(t, client.unhealthyNodes, 1)
	assert.Equal(t, StatusUnhealthy, client.unhealthyNodes[0].Status)
	assert.Equal(t, 0, client.unhealthyNodes[0].Weight)
}

// TestUpdateNodeStatus_Timeout 测试超时时的节点状态更新
func TestUpdateNodeStatus_Timeout(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)
	client.AddNode("http://example.com")

	client.UpdateNodeStatus("http://example.com", ErrTimeout)

	assert.Len(t, client.probationNodes, 1)
	assert.Equal(t, StatusProbation, client.probationNodes[0].Status)
	assert.Equal(t, 4, client.probationNodes[0].Weight) // 5 * 9/10 = 4.5, 向下取整为4
}

// TestUpdateNodeStatus_Throttling 测试限流时的节点状态更新
func TestUpdateNodeStatus_Throttling(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)
	client.AddNode("http://example.com")

	client.UpdateNodeStatus("http://example.com", ErrThrottling)

	assert.Len(t, client.probationNodes, 1)
	assert.Equal(t, StatusProbation, client.probationNodes[0].Status)
	assert.Equal(t, 2, client.probationNodes[0].Weight) // 5 / 2 = 2.5, 向下取整为2
}

// TestUpdateNodeStatus_OtherError 测试其他错误时的节点状态更新
func TestUpdateNodeStatus_OtherError(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)
	client.AddNode("http://example.com")

	client.UpdateNodeStatus("http://example.com", errors.New("未知错误"))

	assert.Len(t, client.healthyNodes, 1)
	assert.Equal(t, StatusHealthy, client.healthyNodes[0].Status)
	assert.Equal(t, 5, client.healthyNodes[0].Weight) // 权重不变
}

// TestUpdateNodeStatus_NonexistentNode 测试更新不存在的节点
func TestUpdateNodeStatus_NonexistentNode(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)

	client.UpdateNodeStatus("http://nonexistent.com", nil)

	assert.Len(t, client.healthyNodes, 0)
	assert.Len(t, client.probationNodes, 0)
	assert.Len(t, client.unhealthyNodes, 0)
}

// TestTryRecoverNodes 测试节点恢复功能
func TestTryRecoverNodes(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Millisecond)
	client.AddNode("http://example.com")

	client.UpdateNodeStatus("http://example.com", ErrNetworkFailure)
	assert.Len(t, client.unhealthyNodes, 1)

	time.Sleep(2 * time.Millisecond) // 等待恢复间隔
	client.tryRecoverNodes()

	assert.Len(t, client.unhealthyNodes, 0)
	assert.Len(t, client.probationNodes, 1)
	assert.Equal(t, StatusProbation, client.probationNodes[0].Status)
	assert.Equal(t, client.minWeight, client.probationNodes[0].Weight)
}

// TestClose 测试关闭客户端功能
func TestClose(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Second)

	// 启动一个goroutine，它将被阻塞直到Close被调用
	done := make(chan bool)
	go func() {
		<-client.stopChan
		done <- true
	}()

	client.Close()

	// 等待goroutine完成或超时
	select {
	case <-done:
		// 测试通过
	case <-time.After(time.Second):
		t.Fatal("Close没有及时停止恢复循环")
	}
}

// TestConcurrentOperations 测试并发操作
func TestConcurrentOperations(t *testing.T) {
	lb := new(MockLoadBalancer)
	client, _ := NewClient(1, 10, 5, lb, time.Millisecond)

	// 添加一些初始节点
	for i := 0; i < 10; i++ {
		client.AddNode(fmt.Sprintf("http://example%d.com", i))
	}

	// 模拟负载均衡器的Select方法
	lb.On("Select", mock.Anything).Return(client.healthyNodes[0], nil)

	var wg sync.WaitGroup
	operations := 1000

	// 并发执行AddNode
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < operations; i++ {
			client.AddNode(fmt.Sprintf("http://newexample%d.com", i))
		}
	}()

	// 并发执行GetNode
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < operations; i++ {
			_, err := client.GetNode()
			assert.NoError(t, err)
		}
	}()

	// 并发执行UpdateNodeStatus
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < operations; i++ {
			client.UpdateNodeStatus("http://example0.com", nil)
		}
	}()

	// 等待所有操作完成
	wg.Wait()

	// 验证最终状态
	assert.Greater(t, len(client.healthyNodes), 10)
	assert.NotPanics(t, func() { client.Close() })
}

package v4

import (
	"errors"
	"log"
	"sync"
	"time"
)

// 节点状态常量
const (
	StatusHealthy   = "healthy"   // 健康状态
	StatusUnhealthy = "unhealthy" // 不健康状态
	StatusProbation = "probation" // 观察状态
)

// 错误类型
var (
	ErrNoAvailableNodes = errors.New("没有可用的节点")
	ErrNetworkFailure   = errors.New("网络故障")
	ErrTimeout          = errors.New("请求超时")
	ErrThrottling       = errors.New("服务限流")
	ErrCircuitBreaker   = errors.New("熔断器开启")
)

// LoadBalancer 定义负载均衡器接口
type LoadBalancer interface {
	Select([]*Node) (*Node, error)
}

// Node 表示一个服务节点
type Node struct {
	URL         string    // 节点URL
	Weight      int       // 节点权重
	Status      string    // 节点状态
	LastCheckAt time.Time // 上次检查时间
}

// Client 表示负载均衡客户端
type Client struct {
	healthyNodes     []*Node
	probationNodes   []*Node
	unhealthyNodes   []*Node
	minWeight        int
	maxWeight        int
	defaultWeight    int
	loadBalancer     LoadBalancer
	mu               sync.RWMutex
	recoveryInterval time.Duration
	stopChan         chan struct{} // 用于停止后台恢复协程
}

// NewClient 创建一个新的客户端实例
func NewClient(minWeight, maxWeight, defaultWeight int, lb LoadBalancer, recoveryInterval time.Duration) (*Client, error) {
	if minWeight < 0 || maxWeight <= minWeight || defaultWeight < minWeight || defaultWeight > maxWeight {
		return nil, errors.New("无效的权重配置")
	}
	c := &Client{
		minWeight:        minWeight,
		maxWeight:        maxWeight,
		defaultWeight:    defaultWeight,
		loadBalancer:     lb,
		recoveryInterval: recoveryInterval,
		stopChan:         make(chan struct{}),
	}
	go c.recoveryLoop()
	return c, nil
}

// recoveryLoop 后台恢复循环
func (c *Client) recoveryLoop() {
	ticker := time.NewTicker(c.recoveryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.tryRecoverNodes()
		case <-c.stopChan:
			return
		}
	}
}

// Close 停止客户端和后台恢复协程
func (c *Client) Close() {
	close(c.stopChan)
}

// AddNode 添加一个新的服务节点
func (c *Client) AddNode(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	node := &Node{URL: url, Weight: c.defaultWeight, Status: StatusHealthy, LastCheckAt: time.Now()}
	c.healthyNodes = append(c.healthyNodes, node)
}

// GetNode 获取一个可用的服务节点
func (c *Client) GetNode() (*Node, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	availableNodes := c.getAvailableNodes()
	if len(availableNodes) == 0 {
		return nil, ErrNoAvailableNodes
	}
	return c.loadBalancer.Select(availableNodes)
}

// getAvailableNodes 获取所有可用的节点，并进行惰性恢复检查
func (c *Client) getAvailableNodes() []*Node {
	// 获取当前时间，用于比较节点的最后检查时间
	now := time.Now()
	// 创建一个切片来存储所有可用的节点
	var availableNodes []*Node

	// 定义一个内部函数，用于检查和追加节点
	checkAndAppend := func(nodes []*Node) {
		for _, node := range nodes {
			// 检查不健康节点是否可以恢复
			if node.Status == StatusUnhealthy && now.Sub(node.LastCheckAt) >= c.recoveryInterval {
				// 将不健康节点移动到试用状态
				c.moveNode(node, StatusProbation)
				// 重置节点权重为最小值
				node.Weight = c.minWeight
				// 更新节点的最后检查时间
				node.LastCheckAt = now
			}
			// 如果节点状态不是 StatusUnhealthy，则认为它是可用的
			if node.Status != StatusUnhealthy {
				availableNodes = append(availableNodes, node)
			}
		}
	}

	// 检查并追加健康节点
	checkAndAppend(c.healthyNodes)
	// 检查并追加试用节点
	checkAndAppend(c.probationNodes)

	// 返回所有可用节点
	return availableNodes
}

// UpdateNodeStatus 更新节点状态和权重
func (c *Client) UpdateNodeStatus(url string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node := c.findNode(url)
	if node == nil {
		log.Printf("未知节点: %s\n", url)
		return
	}

	node.LastCheckAt = time.Now()

	if err == nil {
		c.moveNode(node, StatusHealthy)
		node.Weight = min(node.Weight+1, c.maxWeight)
	} else if errors.Is(err, ErrNetworkFailure) || errors.Is(err, ErrCircuitBreaker) {
		c.moveNode(node, StatusUnhealthy)
		node.Weight = c.minWeight - 1
	} else if errors.Is(err, ErrTimeout) {
		c.moveNode(node, StatusProbation)
		node.Weight = max(node.Weight*9/10, c.minWeight) // 减少10%
	} else if errors.Is(err, ErrThrottling) {
		c.moveNode(node, StatusProbation)
		node.Weight = max(node.Weight/2, c.minWeight) // 减半
	}
}

// findNode 查找指定URL的节点
func (c *Client) findNode(url string) *Node {
	for _, node := range c.healthyNodes {
		if node.URL == url {
			return node
		}
	}
	for _, node := range c.probationNodes {
		if node.URL == url {
			return node
		}
	}
	for _, node := range c.unhealthyNodes {
		if node.URL == url {
			return node
		}
	}
	return nil
}

// moveNode 将节点移动到指定状态
func (c *Client) moveNode(node *Node, toStatus string) {
	if node.Status == toStatus {
		return
	}

	// 从原列表中移除
	var fromList *[]*Node
	switch node.Status {
	case StatusHealthy:
		fromList = &c.healthyNodes
	case StatusProbation:
		fromList = &c.probationNodes
	case StatusUnhealthy:
		fromList = &c.unhealthyNodes
	default:
		panic("不该触发,节点状态错误")
	}

	for i, n := range *fromList {
		if n == node {
			*fromList = append((*fromList)[:i], (*fromList)[i+1:]...)
			break
		}
	}

	// 添加到新列表
	var toList *[]*Node
	switch toStatus {
	case StatusHealthy:
		toList = &c.healthyNodes
	case StatusProbation:
		toList = &c.probationNodes
	case StatusUnhealthy:
		toList = &c.unhealthyNodes
	default:
		panic("不该触发,节点状态错误")
	}

	*toList = append(*toList, node)
	node.Status = toStatus
}

// tryRecoverNodes 尝试恢复不健康的节点
func (c *Client) tryRecoverNodes() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for i := 0; i < len(c.unhealthyNodes); {
		node := c.unhealthyNodes[i]
		if now.Sub(node.LastCheckAt) >= c.recoveryInterval {
			node.Weight = c.minWeight
			c.moveNode(node, StatusProbation)
			node.LastCheckAt = now
		} else {
			i++
		}
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

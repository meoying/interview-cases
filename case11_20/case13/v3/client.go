package v3

import (
	"errors"
	"math"
	"sync"
)

// 定义权重的上下限
const (
	MinWeight = 1
	MaxWeight = 100
)

// 定义可能发生的错误类型
var (
	ErrNetworkFailure = errors.New("网络异常")
	ErrTimeout        = errors.New("请求超时")
	ErrThrottling     = errors.New("服务限流")
	ErrCircuitBreaker = errors.New("服务熔断")
)

// ServiceNode 表示一个服务节点
type ServiceNode struct {
	URL    string // 服务节点的URL
	Weight int    // 服务节点的权重
}

// Client 表示负载均衡客户端
type Client struct {
	nodes sync.Map // 存储服务节点的并发安全map
}

// NewClient 创建一个新的客户端实例
func NewClient() *Client {
	return &Client{}
}

// AddNode 添加一个新的服务节点
func (c *Client) AddNode(url string, weight int) {
	c.nodes.Store(url, &ServiceNode{URL: url, Weight: weight})
}

// AdjustWeight 根据错误类型调整节点权重
func (c *Client) AdjustWeight(url string, err error) {
	if node, ok := c.nodes.Load(url); ok {
		serviceNode := node.(*ServiceNode)
		oldWeight := serviceNode.Weight
		switch {
		case errors.Is(err, ErrNetworkFailure), errors.Is(err, ErrCircuitBreaker):
			// 网络异常或熔断状态，将权重设为0
			serviceNode.Weight = 0
		case errors.Is(err, ErrTimeout):
			// 超时错误，权重减1，但不低于最小权重
			serviceNode.Weight = int(math.Max(float64(oldWeight-1), MinWeight))
		case errors.Is(err, ErrThrottling):
			// 限流状态，将权重设为当前权重的一半，但不低于最小权重
			serviceNode.Weight = int(math.Max(float64(oldWeight)/2, MinWeight))
		}
		c.nodes.Store(url, serviceNode)
	}
}

// GetWeight 获取指定节点的权重
func (c *Client) GetWeight(url string) (int, bool) {
	if node, ok := c.nodes.Load(url); ok {
		return node.(*ServiceNode).Weight, true
	}
	return 0, false
}

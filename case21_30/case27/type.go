package case27

import (
	"errors"
)

// ServiceNode 表示一个服务节点
type ServiceNode struct {
	URL       string // 服务节点的URL
	Weight    int
	CurWeight int
}

var (
	ErrNoAvailableNodes = errors.New("没有可用的节点")
)

// LoadBalancer 定义负载均衡器接口
type LoadBalancer interface {
	Select(nodes []*ServiceNode) (*ServiceNode, error)
}



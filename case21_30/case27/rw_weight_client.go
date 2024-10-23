package case27

import (
	"context"
	"sync"
)

// 定义权重的上下限
const (
	RequestType = "requestType" //
)

type RWWeightClient struct {
	mu         *sync.RWMutex
	writeNodes []*ServiceNode
	readNodes  []*ServiceNode
	// 读节点的负载均衡策略
	readLoadBalancer LoadBalancer
	// 写节点的负载均衡策略
	writeLoadBalancer LoadBalancer
}

func NewRWWeightClient(readLoadBalancer, writeLoadBalancer LoadBalancer) *RWWeightClient {
	return &RWWeightClient{
		mu:                &sync.RWMutex{},
		writeNodes:        make([]*ServiceNode, 0),
		readNodes:         make([]*ServiceNode, 0),
		readLoadBalancer:  readLoadBalancer,
		writeLoadBalancer: writeLoadBalancer,
	}
}

func (r *RWWeightClient) Get(ctx context.Context) (*ServiceNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.isWrite(ctx) {
		return r.writeLoadBalancer.Select(r.writeNodes)
	} else {
		return r.readLoadBalancer.Select(r.readNodes)
	}
}

func (r *RWWeightClient) AddReadNode(node *ServiceNode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.readNodes = append(r.readNodes, node)
	return nil
}

func (r *RWWeightClient) AddWriteNode(node *ServiceNode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.writeNodes = append(r.writeNodes, node)
	return nil
}

func (r *RWWeightClient) isWrite(ctx context.Context) bool {
	val := ctx.Value(RequestType)
	if val == nil {
		return false
	}
	vv, ok := val.(int)
	if !ok {
		return false
	}
	return vv == 1
}

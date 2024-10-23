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
	mu           *sync.RWMutex
	writeNodes   []*ServiceNode
	readNodes    []*ServiceNode
	loadBalancer LoadBalancer
}

func NewRWWeightClient(loadBalancer LoadBalancer) *RWWeightClient {
	return &RWWeightClient{
		mu:           &sync.RWMutex{},
		writeNodes:   make([]*ServiceNode, 0),
		readNodes:    make([]*ServiceNode, 0),
		loadBalancer: loadBalancer,
	}
}

func (r *RWWeightClient) Get(ctx context.Context) (*ServiceNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := r.readNodes
	if r.isWrite(ctx) {
		nodes = r.writeNodes
	}
	bestNode, err := r.loadBalancer.Select(nodes)
	if err != nil {
		return nil, err
	}
	return bestNode, nil
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

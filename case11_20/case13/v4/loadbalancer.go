package v4

// WeightedRoundRobinLoadBalancer 实现了加权轮询算法
type WeightedRoundRobinLoadBalancer struct {
	weights        []int
	currentWeights []int
	lastIndex      int
}

func (lb *WeightedRoundRobinLoadBalancer) Select(nodes []*Node) (*Node, error) {
	if len(nodes) == 0 {
		return nil, ErrNoAvailableNodes
	}

	// 初始化或重置权重
	if len(lb.weights) != len(nodes) {
		lb.weights = make([]int, len(nodes))
		lb.currentWeights = make([]int, len(nodes))
		for i, node := range nodes {
			lb.weights[i] = node.Weight
		}
		lb.lastIndex = -1
	}

	totalWeight := 0
	for i, weight := range lb.weights {
		totalWeight += weight
		lb.currentWeights[i] += weight
	}

	if totalWeight == 0 {
		return nil, ErrNoAvailableNodes
	}

	maxWeight := -1000000
	maxWeightIndex := -1
	for i, weight := range lb.currentWeights {
		if weight > maxWeight {
			maxWeight = weight
			maxWeightIndex = i
		}
	}

	lb.currentWeights[maxWeightIndex] -= totalWeight
	lb.lastIndex = maxWeightIndex

	return nodes[maxWeightIndex], nil
}

package case_648

import (
	"math"
	"sync"
)

const DefaultHashRingSlotNum = 1024

type HashRing struct {
	nodes      []*Node
	slotNodes  []*Node
	requestNum []int
	slotNum    int
	nodeNum    int
	lock       sync.RWMutex
}

func NewHashRing(nodes []*Node, slotNum int) *HashRing {
	cnt := len(nodes)
	avg := slotNum / cnt
	ns := make([]*Node, 0, slotNum)
	k := 0
	total := 0
	for i := 0; i < cnt; i++ {
		if i == cnt-1 {
			total = slotNum
		} else {
			total += avg
		}
		for ; k < total; k++ {
			ns = append(ns, nodes[i])
		}

	}
	return &HashRing{
		nodes:      nodes,
		slotNodes:  ns,
		requestNum: make([]int, slotNum),
		slotNum:    slotNum,
		nodeNum:    len(nodes),
	}
}

func (h *HashRing) GetNode(uid int) *Node {
	sKey := h.getSlotKey(uid)
	h.countRequest(sKey)
	return h.slotNodes[sKey]
}

func (h *HashRing) countRequest(sKey int) {
	h.requestNum[sKey]++
}

func (h *HashRing) getSlotKey(uid int) int {
	return uid % h.slotNum
}

func (h *HashRing) Balance() {
	h.lock.Lock()
	// 统计每节点的槽数和访问数
	totalRequest := 0
	requestNum4Node := make(map[string]int, h.nodeNum)
	slotNum4Node := make(map[string]int, h.nodeNum)
	for k, v := range h.requestNum {
		n := h.slotNodes[k]
		_, ok := requestNum4Node[n.name]
		if !ok {
			requestNum4Node[n.name] = 0
		}
		requestNum4Node[n.name] += v
		totalRequest += v

		_, ok = slotNum4Node[n.name]
		if !ok {
			slotNum4Node[n.name] = 0
		}
		slotNum4Node[n.name]++
	}

	// 算出每节点的每个槽平均访问数
	avgRequest := totalRequest / h.nodeNum
	avgRequest4Node := make(map[string]float64, h.nodeNum)
	for k, v := range requestNum4Node {
		avg := float64(v) / float64(slotNum4Node[k])
		if avg == 0 {
			avgRequest4Node[k] = 1
		} else {
			avgRequest4Node[k] = avg
		}
	}

	// 总的平均节点访问数 / 节点的槽的平均个访问数 = 节点需要多少个槽

	needSlot4Node := make(map[string]int, h.nodeNum)
	totalNeedSlot := 0

	for _, v := range h.nodes {
		_, ok := needSlot4Node[v.name]
		if !ok {
			needSlot4Node[v.name] = 0
		}
		need := int(math.Round(float64(avgRequest) / avgRequest4Node[v.name]))
		needSlot4Node[v.name] = need
		totalNeedSlot += need
	}

	ns := make([]*Node, 0, h.slotNum)
	k := 0
	cnt := 0
	for i := 0; i < h.nodeNum; i++ {
		if i == h.nodeNum-1 {
			cnt = h.slotNum
		} else {
			cnt += int(float64(needSlot4Node[h.nodes[i].name]) / float64(totalNeedSlot) * float64(h.slotNum))
		}
		for ; k < cnt; k++ {
			ns = append(ns, h.nodes[i])
		}
	}

	h.slotNodes = ns
	h.requestNum = make([]int, h.slotNum)
	h.lock.Unlock()
}

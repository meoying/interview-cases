package case12

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHashRing_Balance(t *testing.T) {
	nodes := []*Node{
		{
			name:    "a",
			address: "a_address",
		},
		{
			name:    "b",
			address: "b_address",
		},
		{
			name:    "c",
			address: "c_address",
		},
	}

	testCases := []struct {
		name    string
		uid     int
		before  func() *HashRing
		wantRes string
	}{
		{
			name: "原在A节点，调整后还在A节点",
			uid:  100,
			before: func() *HashRing {
				h := NewHashRing(nodes, DefaultHashRingSlotNum, func(req any) int {
					return req.(int) % DefaultHashRingSlotNum
				})
				return h
			},
			wantRes: "节点a缓存",
		},
		{
			name: "原在B节点，调整后在A节点",
			uid:  4,
			before: func() *HashRing {
				h := NewHashRing(nodes, 10, func(req any) int {
					return req.(int) % 10
				})

				h.SetRequestNumOfSlot([]int{
					10, 20, 30, 40, 50, 60, 70, 80, 90, 100,
				})

				return h
			},
			wantRes: "节点a缓存",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := tc.before()
			n := h.GetNode(tc.uid)
			_, err := n.GetCache(tc.uid)
			require.NoError(t, err)

			h.Balance()

			n = h.GetNode(tc.uid)
			c, err := n.GetCache(tc.uid)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, c)
		})
	}
}

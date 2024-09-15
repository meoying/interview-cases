package case_648

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
				h := NewHashRing(nodes, DefaultHashRingSlotNum)
				return h
			},
			wantRes: "节点a缓存",
		},
		{
			name: "原在A节点，调整后在B节点",
			uid:  100,
			before: func() *HashRing {
				h := NewHashRing(nodes, DefaultHashRingSlotNum)
				uid := 2
				for i := 0; i < 6999; i++ {
					n := h.GetNode(uid)
					n.GetCache(uid)
				}
				uid = 600
				for i := 0; i < 2000; i++ {
					n := h.GetNode(uid)
					n.GetCache(uid)
				}
				uid = 1000
				for i := 0; i < 1000; i++ {
					n := h.GetNode(uid)
					n.GetCache(uid)
				}
				return h
			},
			wantRes: "节点b缓存",
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

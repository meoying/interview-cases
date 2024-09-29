package dao

import (
	"context"
	"fmt"
	"interview-cases/case21_30/case21/domain"
	"sort"
)

// 从db获取数据
type RankDAO interface {
	TopN(ctx context.Context, n int) ([]domain.RankItem, error)
}

// 模拟从数据库拿数据
type MockRankDAO struct {
	Start int
}

func NewMockRankDAO(start int) RankDAO {
	return &MockRankDAO{
		Start: start,
	}
}

func (m *MockRankDAO) TopN(ctx context.Context, n int) ([]domain.RankItem, error) {
	items := make([]domain.RankItem, 0, n)
	for i := 0; i < n; i++ {
		index := m.Start + i
		items = append(items, domain.RankItem{
			Name:  fmt.Sprintf("item_%d", index),
			Score: i,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})
	return items, nil
}

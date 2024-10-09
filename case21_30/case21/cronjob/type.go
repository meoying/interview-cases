package cronjob

import (
	"context"
	"interview-cases/case21_30/case21/domain"
	"sort"
)

// TriSvc 第三方服务用于模拟前1000名的数据
type TriSvc interface {
	TopN(ctx context.Context, n int) ([]domain.RankItem, error)
}

type triSvc struct {
	start int
}

func NewMockTriSvc(start int) TriSvc {
	return &triSvc{start: start}
}

func (t *triSvc) TopN(ctx context.Context, n int) ([]domain.RankItem, error) {
	items := make([]domain.RankItem, 0, n)
	for i := 0; i < n; i++ {
		index := t.start + i
		items = append(items, domain.RankItem{
			ID:  int64( index),
			Score: i,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})
	return items, nil
}

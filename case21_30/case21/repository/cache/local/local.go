package local

import (
	"context"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"interview-cases/case21_30/case21/domain"
)

type Cache struct {
	topN *atomicx.Value[[]domain.RankItem]
}

func NewCache() *Cache {
	return &Cache{
		topN: atomicx.NewValue[[]domain.RankItem](),
	}
}

func (c *Cache) Set(ctx context.Context, rankItem []domain.RankItem) error {
	c.topN.Store(rankItem)
	return nil
}

func (c *Cache) Get(ctx context.Context) ([]domain.RankItem, error) {
	items := c.topN.Load()
	return items, nil
}

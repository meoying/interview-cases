package local

import (
	"context"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"interview-cases/case21_30/case21/domain"
)

type Cache struct {
	topN *atomicx.Value[[]domain.Article]
}

func NewCache() *Cache {
	return &Cache{
		topN: atomicx.NewValue[[]domain.Article](),
	}
}

func (c *Cache) Set(ctx context.Context, rankItem []domain.Article) error {
	c.topN.Store(rankItem)
	return nil
}

func (c *Cache) Get(ctx context.Context) ([]domain.Article, error) {
	items := c.topN.Load()
	return items, nil
}

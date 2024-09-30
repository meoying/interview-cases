package redis

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/redis/go-redis/v9"
	"interview-cases/case21_30/case21/domain"
)

type Cache struct {
	client redis.Cmdable
	key    string
}

func NewCache(client redis.Cmdable, key string) *Cache {
	return &Cache{
		client: client,
		key:    key,
	}
}

func (c *Cache) Set(ctx context.Context, rankItem []domain.RankItem) error {
	members := slice.Map(rankItem, func(idx int, src domain.RankItem) redis.Z {
		return redis.Z{Score: float64(src.Score), Member: src.Name}
	})
	err := c.client.ZAdd(ctx, c.key, members...).Err()
	return err
}

func (c *Cache) Get(ctx context.Context, n int) ([]domain.RankItem, error) {
	members, err := c.client.ZRevRangeWithScores(ctx, c.key, 0, int64(n)).Result()
	if err != nil {
		return nil, err
	}
	list := slice.Map(members, func(idx int, src redis.Z) domain.RankItem {
		return domain.RankItem{
			Name:  src.Member.(string),
			Score: int(src.Score),
		}
	})
	return list, nil
}

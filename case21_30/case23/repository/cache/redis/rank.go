package redis

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/redis/go-redis/v9"
	"interview-cases/case21_30/case23/domain"
	"interview-cases/case21_30/case23/repository/cache/redis/mergesort"
	"strconv"
)

const (
	defaultKeyNumber = 10 // 分key分几个
)

type Cache struct {
	client    redis.Cmdable
	key       string
	keyNumber int
}

func NewCache(client redis.Cmdable) *Cache {
	return &Cache{
		client:    client,
		key:       "rank",
		keyNumber: defaultKeyNumber,
	}
}

func (r *Cache) Set(ctx context.Context, item domain.RankItem) error {
	return r.client.ZAdd(ctx, fmt.Sprintf("%s:%d", r.key, item.ID%int64(r.keyNumber)), redis.Z{
		Score:  float64(item.Score),
		Member: item.ID,
	}).Err()
}

func (r *Cache) Get(ctx context.Context, n int) ([]domain.RankItem, error) {
	rankLists := make([][]domain.RankItem, 0, n)
	for i := 0; i < r.keyNumber; i++ {
		key := fmt.Sprintf("%s:%d", r.key, i)
		members, err := r.client.ZRevRangeWithScores(ctx, key, 0, int64(n)).Result()
		if err != nil {
			return nil, err
		}
		list := slice.Map(members, func(idx int, src redis.Z) domain.RankItem {
			mem := src.Member
			id, _ := strconv.ParseInt(mem.(string), 10, 64)
			return domain.RankItem{
				ID:    id,
				Score: int64(src.Score),
			}
		})
		rankLists = append(rankLists, list)
	}
	// 归并排序
	return mergesort.GetSortList(rankLists, n), nil
}

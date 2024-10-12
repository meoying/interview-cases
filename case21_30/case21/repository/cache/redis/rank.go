package redis

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/redis/go-redis/v9"
	"interview-cases/case21_30/case21/domain"
	"strconv"
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

func (c *Cache) Set(ctx context.Context, rankItem []domain.Article) error {
	members := slice.Map(rankItem, func(idx int, src domain.Article) redis.Z {
		return redis.Z{Score: float64(src.LikeCnt), Member: src.ID}
	})
	err := c.client.ZAddArgs(ctx, c.key, redis.ZAddArgs{
		Members: members,
		// 保证不增加额外元素
		XX: true,
		// 只会在新分数大于当前分数的情况下更新
		GT: true,
	}).Err()
	return err
}

func (c *Cache) Get(ctx context.Context, n int) ([]domain.Article, error) {
	members, err := c.client.ZRevRangeWithScores(ctx, c.key, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	list := slice.Map(members, func(idx int, src redis.Z) domain.Article {
		mem := src.Member.(string)
		id, _ := strconv.ParseInt(mem, 10, 64)
		return domain.Article{
			ID:      id,
			LikeCnt: int(src.Score),
		}
	})
	return list, nil
}

// SyncRank 用于定时任务同步到redis，先删除然后将数据重新写入
func (c *Cache) SyncRank(ctx context.Context, rankItems []domain.Article) error {
	mems := slice.Map(rankItems, func(idx int, src domain.Article) redis.Z {
		return redis.Z{
			Score:  float64(src.LikeCnt),
			Member: src.ID,
		}
	})
	pipe := c.client.TxPipeline()
	pipe.Del(ctx, c.key)
	pipe.ZAdd(ctx, c.key, mems...)
	res, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	if res[0].Err() != nil {
		return fmt.Errorf("删除原有的键失败 %v", res[0].Err())
	}
	if res[1].Err() != nil {
		return fmt.Errorf("往zset添加元素失败 %v", res[1].Err())
	}
	return nil
}

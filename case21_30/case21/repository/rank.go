package repository

import (
	"context"
	"interview-cases/case21_30/case21/domain"
	"interview-cases/case21_30/case21/repository/cache/local"
	"interview-cases/case21_30/case21/repository/cache/redis"
)

type RankRepository interface {
	TopN(ctx context.Context) ([]domain.Article, error)
	ReplaceTopN(ctx context.Context, items []domain.Article) error
}

func NewRankRepository(localCache *local.Cache, redisCache *redis.Cache) RankRepository {
	return &rankRepository{
		localCache: localCache,
		redisCache: redisCache,
	}
}

type rankRepository struct {
	localCache *local.Cache
	redisCache *redis.Cache
}

func (r *rankRepository) TopN(ctx context.Context) ([]domain.Article, error) {
	// 直接读取本地
	return r.localCache.Get(ctx)
}

func (r *rankRepository) ReplaceTopN(ctx context.Context, items []domain.Article) error {
	// 写入redis
	return r.redisCache.Set(ctx, items)
}

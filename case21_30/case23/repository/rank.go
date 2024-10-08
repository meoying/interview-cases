package repository

import (
	"context"
	"interview-cases/case21_30/case23/domain"
	"interview-cases/case21_30/case23/repository/cache/local"
	"interview-cases/case21_30/case23/repository/cache/redis"
)

type RankRepo interface {
	TopN(ctx context.Context) ([]domain.RankItem, error)
	ReplaceTopN(ctx context.Context, item domain.RankItem) error
}

type rankRepo struct {
	localCache *local.Cache
	redisCache *redis.Cache
}

func NewRankRepo(localCache *local.Cache,redisCache *redis.Cache)RankRepo{
	return &rankRepo{
		localCache: localCache,
		redisCache: redisCache,
	}
}

func (r *rankRepo) TopN(ctx context.Context) ([]domain.RankItem, error) {
	// 从本地缓存获取排名
	return r.localCache.Get(ctx)
}

func (r *rankRepo) ReplaceTopN(ctx context.Context, item domain.RankItem) error {
	// 设置缓存
	return r.redisCache.Set(ctx,item)
}

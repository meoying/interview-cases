package service

import (
	"context"
	"interview-cases/case21_30/case21/domain"
	"interview-cases/case21_30/case21/repository"
)


// 榜单服务
type RankService interface {
	// 榜单前一百
	TopN(ctx context.Context) (items []domain.RankItem, err error)
	// 更新榜单数据
	Update(ctx context.Context, items []domain.RankItem) (err error)
}


type topSvc struct {
	repo repository.RankRepository
}
func NewRankService(repo repository.RankRepository) RankService {
	return &topSvc{
		repo: repo,
	}
}

func (t *topSvc) TopN(ctx context.Context) ([]domain.RankItem, error) {
	return t.repo.TopN(ctx)
}

func (t *topSvc) Update(ctx context.Context, item []domain.RankItem) (err error) {
	return t.repo.ReplaceTopN(ctx,item)
}

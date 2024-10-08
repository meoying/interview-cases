package service

import (
	"context"
	"interview-cases/case21_30/case23/domain"
	"interview-cases/case21_30/case23/repository"
)

type RankService interface {
	TopN(ctx context.Context) ([]domain.RankItem, error)
	UpdateScore(ctx context.Context, rankItem domain.RankItem) error
}

type rankSvc struct {
	repo repository.RankRepo
}

func (r *rankSvc) TopN(ctx context.Context) ([]domain.RankItem, error) {
	return r.repo.TopN(ctx)
}

func (r *rankSvc) UpdateScore(ctx context.Context, rankItems domain.RankItem) error {
	return r.repo.ReplaceTopN(ctx, rankItems)
}

func NewRankSvc(repo repository.RankRepo) RankService {
	return &rankSvc{repo: repo}
}

package service

import (
	"context"
	"interview-cases/case21_30/case21/cronjob/tri/domain"
	"interview-cases/case21_30/case21/cronjob/tri/repository"
)

type ArticleSvc interface {
	// BatchCreate 批量创建
	BatchCreate(ctx context.Context, articles []domain.Article) error
	// TopN 获取前1000名
	TopN(ctx context.Context,n int) ([]domain.Article, error)
}

type articleSvc struct {
	repo repository.ArticleRepo
}
func NewArticleSvc(repo repository.ArticleRepo)ArticleSvc{
	return &articleSvc{
		repo: repo,
	}
}

func (a *articleSvc) BatchCreate(ctx context.Context, articles []domain.Article) error {
	return a.repo.BatchCreate(ctx,articles)
}

func (a *articleSvc) TopN(ctx context.Context, n int) ([]domain.Article, error) {
	return a.repo.TopN(ctx,n)
}


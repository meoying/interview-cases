package service

import (
	"context"
	"interview-cases/case11_20/case14/domain"
	"interview-cases/case11_20/case14/repository"
)

type DelayMsgSvc interface {
	Insert(ctx context.Context, msg domain.DelayMsg) error
	// 批量更新成完成
	Complete(ctx context.Context, ids []int64) error
	FindDelayMsg(ctx context.Context) ([]domain.DelayMsg, error)
}


type delayMsgSvc struct {
	repo repository.DelayMsgRepository
}

func NewDelayMsgSvc(repo repository.DelayMsgRepository) DelayMsgSvc {
	return &delayMsgSvc{
		repo: repo,
	}
}

func (d *delayMsgSvc) Insert(ctx context.Context, msg domain.DelayMsg) error {
	return d.repo.Insert(ctx, msg)
}

func (d *delayMsgSvc) Complete(ctx context.Context, ids []int64) error {
	return d.repo.Complete(ctx, ids)
}

func (d *delayMsgSvc) FindDelayMsg(ctx context.Context) ([]domain.DelayMsg, error) {
	return d.repo.FindDelayMsg(ctx)
}

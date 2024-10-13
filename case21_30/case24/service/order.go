package service

import (
	"context"
	"interview-cases/case21_30/case24/domain"
	"interview-cases/case21_30/case24/repository"
)

type OrderService interface {
	Save(ctx context.Context, order domain.Order) error
	Get(ctx context.Context, id int64) (domain.Order, error)
}

func NewOrderService(repo repository.OrderRepo)OrderService{
	return &orderService{
		repo: repo,
	}
}

type orderService struct {
	repo repository.OrderRepo
}

func (o *orderService) Save(ctx context.Context, order domain.Order) error {
	return o.repo.Save(ctx, order)
}

func (o *orderService) Get(ctx context.Context, id int64) (domain.Order, error) {
	return o.repo.Get(ctx,id)
}

package repository

import (
	"context"
	"errors"
	"interview-cases/case21_30/case24/domain"
	"interview-cases/case21_30/case24/repository/cache"
	"interview-cases/case21_30/case24/repository/dao"
	"log/slog"
)

type OrderRepo interface {
	Save(ctx context.Context, order domain.Order) error
	Get(ctx context.Context, id int64) (domain.Order, error)
}

func NewOrderRepo(orderDao   dao.OrderDAO,orderCache cache.Cache)OrderRepo {
	return &orderRepo{
		orderDao: orderDao,
		orderCache: orderCache,
	}
}

type orderRepo struct {
	orderDao   dao.OrderDAO
	orderCache cache.Cache
}

func (o *orderRepo) Save(ctx context.Context, order domain.Order) error {
	err := o.orderDao.Save(ctx, dao.Order{
		ID:      order.ID,
		Name:    order.Name,
		BuyerID: order.BuyerID,
		Price:   order.Price,
	})
	if err != nil {
		return err
	}
	if order.ID != 0 {
		err = o.orderCache.Del(ctx, order.ID)
		if err != nil {
			slog.Error("删除缓存失败", slog.Any("err", err))
		}
	}
	return nil
}

func (o *orderRepo) Get(ctx context.Context, id int64) (domain.Order, error) {
	// 先从缓存中获取
	order, err := o.orderCache.Get(ctx, id)
	if err == nil {
		return order, nil
	}
	if !errors.Is(err, cache.ErrKeyNotFound) {
		slog.Error("从缓存中获取数据失败", slog.Any("err", err))
	}
	orderEntity, err := o.orderDao.Get(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	// 设置缓存
	order = domain.Order{
		ID:      orderEntity.ID,
		Name:    orderEntity.Name,
		BuyerID: orderEntity.BuyerID,
		Price:   orderEntity.Price,
	}
	err = o.orderCache.Set(ctx, order)
	if err != nil {
		slog.Error("设置缓存失败", slog.Any("err", err))
	}
	return order, nil
}

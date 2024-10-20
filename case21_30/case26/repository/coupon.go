package repository

import (
	"context"
	"interview-cases/case21_30/case26/repository/cache"
)

type CouponRepository interface {
	DecrCoupon(ctx context.Context) error
	GetCoupon(ctx context.Context) (int, error)
	CheckUidExist(ctx context.Context, uid int) (bool,error)
}

type couponRepositoryImpl struct {
	cache.Cache
}
func NewCouponRepository(couponCache cache.Cache)CouponRepository {
	return &couponRepositoryImpl{
		Cache: couponCache,
	}
}


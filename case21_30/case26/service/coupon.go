package service

import (
	"context"
	"errors"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"interview-cases/case21_30/case26/repository"
	"interview-cases/case21_30/case26/repository/cache"
	"math/rand/v2"
)

type CouponSvc interface {
	// Preempt 抢优惠券,uid为用户一个用户不能反复抢。 bool抢成功还是失败
	Preempt(ctx context.Context, uid int64) (bool, error)
}

type CouponSvcImpl struct {
	repo repository.CouponRepository
	// 10w库存 50%直接拒绝 1w库存10%直接拒绝
	m atomicx.Value[int]
}

// Preempt
func (c *CouponSvcImpl) Preempt(ctx context.Context, uid int64) (bool, error) {
	ok, err := c.repo.SetUidNX(ctx, uid)
	if err != nil {
		return false, err
	}
	if !ok {
		// 说明已经抢过了
		return false, nil
	}
	if c.randomReject() {
		return false, nil
	}
	// 扣减库存
	err = c.repo.DecrCoupon(ctx)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, cache.ErrInsufficientCoupon):
		return false, nil
	default:
		return false, err
	}
}

func (c *CouponSvcImpl) randomReject() bool {
	randomValue := rand.IntN(100)
	return randomValue > c.m.Load()
}

package service

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"interview-cases/case21_30/case26/repository"
	"interview-cases/case21_30/case26/repository/cache"
	"log"
	"log/slog"
	"math/rand/v2"
	"time"
)

type CouponSvc interface {
	// Preempt 抢优惠券,uid为用户一个用户不能反复抢。 bool抢成功还是失败
	Preempt(ctx context.Context, uid int) (bool, error)
}

const defaultCoupon = 100000

type CouponSvcImpl struct {
	repo repository.CouponRepository
	// 10w库存 50%直接拒绝 没减少1w 减少10%的拒绝率直到为0
	m *atomicx.Value[int]
}

func NewCouponSvc(repo repository.CouponRepository) CouponSvc {
	couponSvc := &CouponSvcImpl{
		repo: repo,
		m:    atomicx.NewValueOf[int](50),
	}
	go couponSvc.adjust()
	return couponSvc
}

// Preempt
func (c *CouponSvcImpl) Preempt(ctx context.Context, uid int) (bool, error) {
	ok, err := c.repo.CheckUidExist(ctx, uid)
	if err != nil {
		return false, err
	}
	if ok {
		// 说明已经抢过了
		log.Println("抢过了")
		return false, nil
	}
	if c.randomReject() {
		log.Println("被随机拒绝了")
		return false, nil
	}
	// 扣减库存
	err = c.repo.DecrCoupon(ctx)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, cache.ErrInsufficientCoupon):
		log.Println("库存不足")
		return false, nil
	default:
		return false, err
	}
}

func (c *CouponSvcImpl) randomReject() bool {
	randomValue := rand.IntN(100)
	return randomValue < c.m.Load()
}

func (c *CouponSvcImpl) adjust() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		coupons, err := c.repo.GetCoupon(ctx)
		cancel()
		if err != nil {
			slog.Error("获取库存失败", slog.Any("err", err))
		}
		reduceCoupons := defaultCoupon - coupons
		wantAdjust := max(0, 50-(reduceCoupons/10000)*10)
		c.m.Store(wantAdjust)
		time.Sleep(10 * time.Millisecond)
	}
}

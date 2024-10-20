package cache

import (
	"context"
	"github.com/pkg/errors"
)

var ErrInsufficientCoupon = errors.New("库存不足")

type Cache interface {
	// DecrCoupon 扣减库存 返回值是扣减完的库存，如果库存不足则返回错误
	DecrCoupon(ctx context.Context) error
	// GetCoupon 获取库存
	GetCoupon(ctx context.Context) (int, error)
	// SetUidNX 设置抽奖的用户 true设置 false加载
	CheckUidExist(ctx context.Context, uid int) (bool,error)
}

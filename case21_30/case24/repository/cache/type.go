package cache

import (
	"context"
	"github.com/pkg/errors"
	"interview-cases/case21_30/case24/domain"
)

var ErrKeyNotFound = errors.New("键不存在")

type Cache interface {
	Get(ctx context.Context, id int64) (domain.Order, error)
	Set(ctx context.Context, val domain.Order) error
	Del(ctx context.Context,id int64)error
}

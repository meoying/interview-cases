package cache

import (
	"context"
	"github.com/pkg/errors"
)

var ErrKeyNotFound = errors.New("键不存在")

type Cache interface {
	Get(ctx context.Context, key string) (any,error)
	Set(ctx context.Context, key string, val any)error
}




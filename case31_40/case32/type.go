package case32

import (
	"context"
	"time"
)

type Strategy interface {
	Next(ctx context.Context, err error) (time.Duration, bool)
}

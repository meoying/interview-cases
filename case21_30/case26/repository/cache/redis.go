package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/redis/go-redis/v9"
	"math/rand/v2"
	"strconv"
	"sync"
)

var (
	//go:embed lua/bloom.lua
	bloomLua    string
	bloomFilter string = "bloom_filter"
)

type RedisCoupon struct {
	client redis.Cmdable
	//key 和 keyNumber从配置文件获取
	key         string
	keyNumber   int
	mu          *sync.RWMutex
	activeSlice []bool
}

func NewRedisCoupon(ctx context.Context, client redis.Cmdable, key string, keyNumber, count int) (*RedisCoupon, error) {
	r := &RedisCoupon{
		client:      client,
		key:         key,
		keyNumber:   keyNumber,
		mu:          &sync.RWMutex{},
		activeSlice: make([]bool, keyNumber),
	}
	err := r.setCoupon(ctx, key, count)
	if err != nil {
		return nil, fmt.Errorf("设置库存失败 %v", err)
	}
	return r, nil
}

func (r *RedisCoupon) GetCoupon(ctx context.Context) (int, error) {
	pipe := r.client.TxPipeline()
	resList := make([]*redis.StringCmd, 0)
	for i := 0; i < r.keyNumber; i++ {
		resList = append(resList, pipe.Get(ctx, r.key))
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	var sum int
	for _, res := range resList {
		v, err := res.Result()
		if err != nil {
			return 0, err
		}
		coupon, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		sum += coupon
	}
	return sum, nil
}

func (r *RedisCoupon) setCoupon(ctx context.Context, key string, count int) error {
	var keysMap = make(map[string]int)
	quotient := count / r.keyNumber
	remainder := count % r.keyNumber
	for i := 0; i < r.keyNumber; i++ {
		if i == r.keyNumber-1 {
			keysMap[fmt.Sprintf("%s:%d", key, i)] = quotient + remainder
		} else {
			keysMap[fmt.Sprintf("%s:%d", key, i)] = quotient
		}
	}
	pipe := r.client.TxPipeline()
	for k, v := range keysMap {
		pipe.Set(ctx, k, v, 0)
	}
	_, err := pipe.Exec(ctx)
	r.activeSlice = slice.Map(r.activeSlice, func(idx int, src bool) bool {
		return true
	})
	return err
}

func (r *RedisCoupon) DecrCoupon(ctx context.Context) error {
	idx := rand.IntN(r.keyNumber)
	for i := 0; i < r.keyNumber; i++ {
		v := idx + i
		if r.getActiveSlice(v % r.keyNumber) {
			continue
		}
		coupon, err := r.client.Decr(ctx, fmt.Sprintf("%s:%d", r.key, v%r.keyNumber)).Result()
		if err != nil {
			return err
		}
		if coupon > 0 {
			return nil
		}
		if coupon == 0 {
			r.setActiveSlice(v % r.keyNumber)
			return nil
		}
	}
	return ErrInsufficientCoupon
}

func (r *RedisCoupon) CheckUidExist(ctx context.Context, uid int64) (bool, error) {
	res, err := r.client.Eval(ctx, bloomLua, []string{bloomFilter}, uid).Result()
	if err != nil {
		 return false, err
	}
	existed := res.(bool)
	return existed, nil
}

func (r *RedisCoupon) setActiveSlice(idx int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.activeSlice[idx] = false
}

func (r *RedisCoupon) getActiveSlice(idx int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.activeSlice[idx]
}

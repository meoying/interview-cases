package case26

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"interview-cases/case21_30/case26/repository"
	"interview-cases/case21_30/case26/repository/cache"
	"interview-cases/case21_30/case26/service"
	"interview-cases/test"
	"log"
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	svc service.CouponSvc
}

func (t *TestSuite) SetupSuite() {
	client := test.InitRedisBloom()
	couponClient, err := cache.NewRedisCoupon(context.Background(), client, "coupon", 10, 100000)
	require.NoError(t.T(), err)
	repo := repository.NewCouponRepository(couponClient)
	couponSvc := service.NewCouponSvc(repo)
	t.svc = couponSvc
}

func (t *TestSuite) TestPreempt() {
	// 测试流程
	// 有 10w库存，100w用户抢占
	// 抢到的人
	preemptUids := make([]int, 0, 100000)
	mu := &sync.RWMutex{}
	wg := &sync.WaitGroup{}
	for i := 0; i <= 500000; i++ {
		// 保证获取连接不超时，可以适当提高睡眠的时间
		time.Sleep(110 * time.Microsecond)
		if rand.IntN(100) > 20 {
			wg.Add(1)
			uid := i
			go func() {
				defer wg.Done()
				ok, err := t.svc.Preempt(context.Background(), uid)
				require.NoError(t.T(), err)
				if ok {
					log.Printf("uid%d 抢到了", uid)
					mu.Lock()
					preemptUids = append(preemptUids, uid)
					mu.Unlock()
				}
			}()
		} else {
			// 20%的几率会抢两次
			wg.Add(2)
			uid := i
			go func() {
				defer wg.Done()
				ok, err := t.svc.Preempt(context.Background(), uid)
				require.NoError(t.T(), err)
				if ok {
					log.Printf("uid%d 抢到了", uid)
					mu.Lock()
					preemptUids = append(preemptUids, uid)
					mu.Unlock()
				}
			}()
			go func() {
				defer wg.Done()
				ok, err := t.svc.Preempt(context.Background(), uid)
				require.NoError(t.T(), err)
				if ok {
					log.Printf("uid%d 抢到了", uid)
					mu.Lock()
					preemptUids = append(preemptUids, uid)
					mu.Unlock()
				}
			}()
		}
	}
	wg.Wait()
	// 校验抢占人数
	require.Equal(t.T(), 100000, len(preemptUids))
	// 校验人都是不同的
	uidMap := map[int]struct{}{}
	for _, uid := range preemptUids {
		_, ok := uidMap[uid]
		require.False(t.T(), ok)
		uidMap[uid] = struct{}{}
	}
}

func TestPreempt(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

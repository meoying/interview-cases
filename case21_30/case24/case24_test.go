package case24

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"interview-cases/case21_30/case24/domain"
	"interview-cases/case21_30/case24/repository"
	"interview-cases/case21_30/case24/repository/cache/local"
	"interview-cases/case21_30/case24/repository/cache/mix"
	"interview-cases/case21_30/case24/repository/cache/redis"
	"interview-cases/case21_30/case24/repository/dao"
	"interview-cases/case21_30/case24/service"
	"interview-cases/test"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	orderSvc   service.OrderService
	localCache *local.Cache
	redisCache *redis.Cache
}

func (t *TestSuite) SetupSuite() {
	db := test.InitDB()
	err := dao.InitTables(db)
	require.NoError(t.T(), err)
	client := test.InitRedis()
	mockClient := NewRedisMock(client)
	localCache := local.NewCache()
	redisCache := redis.NewCache(mockClient)
	ca := mix.NewCache(localCache, redisCache, 0)
	orderDao := dao.NewOrderDAO(db)
	orderRepo := repository.NewOrderRepo(orderDao, ca)
	orderSvc := service.NewOrderService(orderRepo)
	t.localCache = localCache
	t.redisCache = redisCache
	t.orderSvc = orderSvc
}

func (t *TestSuite) TestRedis() {
	t.initOrders()
	// 更新数据
	err := t.orderSvc.Save(context.Background(), domain.Order{
		ID:      1,
		Name:    fmt.Sprintf("order_new_%d", 1),
		BuyerID: 123,
		Price:   777,
	})
	require.NoError(t.T(), err)
	// 查看数据
	order, err := t.orderSvc.Get(context.Background(), 1)
	require.NoError(t.T(), err)
	assert.Equal(t.T(), domain.Order{
		ID:      1,
		Name:    fmt.Sprintf("order_new_%d", 1),
		BuyerID: 123,
		Price:   777,
	}, order)
	// 等待11s，redis崩溃
	time.Sleep(11 * time.Second)
	err = t.orderSvc.Save(context.Background(), domain.Order{
		ID:      1,
		Name:    fmt.Sprintf("order_new_%d", 1),
		BuyerID: 123,
		Price:   666,
	})
	require.NoError(t.T(), err)

	// 直接获取详情数据
	order, err = t.orderSvc.Get(context.Background(), 1)
	require.NoError(t.T(), err)
	assert.Equal(t.T(), domain.Order{
		ID:      1,
		Name:    fmt.Sprintf("order_new_%d", 1),
		BuyerID: 123,
		Price:   666,
	}, order)

	// 查看并校验本地缓存中的数据
	order, err = t.localCache.Get(context.Background(), 1)
	require.NoError(t.T(), err)
	assert.Equal(t.T(), domain.Order{
		ID:      1,
		Name:    fmt.Sprintf("order_new_%d", 1),
		BuyerID: 123,
		Price:   666,
	}, order)

	// 又过10s检测到恢复
	time.Sleep(10 * time.Second)
	err = t.orderSvc.Save(context.Background(), domain.Order{
		ID:      2,
		Name:    fmt.Sprintf("order_new_%d", 2),
		BuyerID: 123,
		Price:   999,
	})
	require.NoError(t.T(), err)
	// 直接获取详情数据
	order, err = t.orderSvc.Get(context.Background(), 2)
	require.NoError(t.T(), err)
	assert.Equal(t.T(), domain.Order{
		ID:      2,
		Name:    fmt.Sprintf("order_new_%d", 2),
		BuyerID: 123,
		Price:   999,
	}, order)
	// 查看并校验redis中的数据
	order, err = t.redisCache.Get(context.Background(), 2)
	require.NoError(t.T(), err)
	assert.Equal(t.T(), domain.Order{
		ID:      2,
		Name:    fmt.Sprintf("order_new_%d", 2),
		BuyerID: 123,
		Price:   999,
	}, order)

}

func (t *TestSuite) initOrders() {
	for i := 1; i <= 10; i++ {
		err := t.orderSvc.Save(context.Background(), domain.Order{
			ID:      int64(i),
			Name:    fmt.Sprintf("order_%d", i),
			BuyerID: 123,
			Price:   int32(i + 10),
		})
		require.NoError(t.T(), err)
	}
	return
}

func TestDelayMsg(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

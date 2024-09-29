package case21

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"interview-cases/case21_30/case21/cronjob"
	"interview-cases/case21_30/case21/repository"
	"interview-cases/case21_30/case21/repository/cache/local"
	"interview-cases/case21_30/case21/repository/cache/redis"
	"interview-cases/case21_30/case21/repository/dao"
	"interview-cases/case21_30/case21/service"
	"interview-cases/test"
	"time"
)

type TestSuite struct {
	suite.Suite
	localCache *local.Cache
	redisCache *redis.Cache
}

func (t *TestSuite) SetupSuite() {
	// 初始化
	t.initServer()
}

func (t *TestSuite) TestRank() {
	// 等待两分钟看定时任务有没有将数据同步到
	time.Sleep(2 * time.Minute)
	// 查看cache中的数据

	//

}

func (t *TestSuite) CheckCacheData() {
	items, err := t.redisCache.Get(context.Background(), 1000)
	require.NoError(t.T(), err)

}

func (t *TestSuite) initServer() {
	client := test.InitRedis()
	localCache := local.NewCache()
	redisCache := redis.NewCache(client, "rank")
	db := dao.NewMockRankDAO(0)
	err := cronjob.InitJob(db, redisCache, localCache)
	require.NoError(t.T(), err)
	repo := repository.NewRankRepository(localCache, redisCache)
	service.NewRankService(repo)
}

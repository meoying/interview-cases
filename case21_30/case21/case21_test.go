package case21

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"interview-cases/case21_30/case21/cronjob"
	"interview-cases/case21_30/case21/domain"
	"interview-cases/case21_30/case21/repository"
	"interview-cases/case21_30/case21/repository/cache/local"
	"interview-cases/case21_30/case21/repository/cache/redis"
	"interview-cases/case21_30/case21/service"
	"interview-cases/test"
	"log"
	"testing"

	"time"
)

type TestSuite struct {
	suite.Suite
	svc        service.RankService
	localCache *local.Cache
	redisCache *redis.Cache
}

func (t *TestSuite) SetupSuite() {
	// 初始化
	t.initServer()
}

func (t *TestSuite) TestRank() {
	// 更新数据
	t.updateItems()
	log.Println("更新数据")
	// 等待一分钟看定时任务有没有将redis数据同步到本地缓存
	log.Println("等待一分钟同步本地缓存。。。")
	time.Sleep(1*time.Minute + 10*time.Second)
	// 查看前100名
	items, err := t.svc.TopN(context.Background())
	require.NoError(t.T(), err)
	// 校验本地缓存中的数据
	t.checkItems(items, t.getData(1901, 2000))
	// 再等两分钟
	log.Println("等待两分钟将全局数据更新进redis。。。")
	time.Sleep(2*time.Minute + 10*time.Second)
	// 校验redis中的数据是否已经同步
	t.checkCacheData()
	log.Println("等待一分钟同步进本地缓存。。。")
	// 再等一分钟等待redis中的数据是否同步到本地缓存
	time.Sleep(1*time.Minute + 10*time.Second)
	// 获取本地缓存中的数据
	items, err = t.svc.TopN(context.Background())
	require.NoError(t.T(), err)
	// 校验本地缓存中的数据
	t.checkItems(items, t.getData(900, 999))
}

// 更新数据
func (t *TestSuite) updateItems() {
	items := t.getData(1000, 2000)
	err := t.svc.Update(context.Background(), items)
	require.NoError(t.T(), err)
}

func (t *TestSuite) checkCacheData() {
	wantItems := t.getData(0, 999)
	actualItems, err := t.redisCache.Get(context.Background(), 1000)
	require.NoError(t.T(), err)
	assert.Equal(t.T(), wantItems, actualItems)
}

// 校验获取数据是否正确
func (t *TestSuite) checkItems(actualItems []domain.RankItem, wantItems []domain.RankItem) {
	assert.Equal(t.T(), wantItems, actualItems)
}

func (t *TestSuite) initServer() {
	client := test.InitRedis()
	localCache := local.NewCache()
	redisCache := redis.NewCache(client, "rank")
	triSvc := cronjob.NewMockTriSvc(0)
	err := cronjob.InitJob(triSvc, redisCache, localCache)
	require.NoError(t.T(), err)
	repo := repository.NewRankRepository(localCache, redisCache)
	t.svc = service.NewRankService(repo)
	t.localCache = localCache
	t.redisCache = redisCache
}

func (t *TestSuite) getData(start, end int) []domain.RankItem {
	items := make([]domain.RankItem, 0)
	for i := end; i >= start; i-- {
		items = append(items, domain.RankItem{
			ID:    int64(i),
			Score: i,
		})
	}
	return items
}

func TestTopN(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

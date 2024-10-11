package case21

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"interview-cases/case21_30/case21/cronjob"
	"interview-cases/case21_30/case21/cronjob/tri"
	triDomain "interview-cases/case21_30/case21/cronjob/tri/domain"
	triSvc "interview-cases/case21_30/case21/cronjob/tri/service"
	"interview-cases/case21_30/case21/domain"
	"interview-cases/case21_30/case21/repository"
	"interview-cases/case21_30/case21/repository/cache/local"
	"interview-cases/case21_30/case21/repository/cache/redis"
	"interview-cases/case21_30/case21/service"
	"interview-cases/test"
	"testing"
)

type TestSuite struct {
	suite.Suite
	articleSvc triSvc.ArticleSvc
	triSvc     cronjob.TriSvc
	svc        service.RankService
	localCache *local.Cache
	redisCache *redis.Cache
}

func (t *TestSuite) SetupSuite() {
	// 初始化
	t.initServer()
}

func (t *TestSuite) TestRank() {
	// 查看前100名
	items, err := t.svc.TopN(context.Background())
	require.NoError(t.T(), err)
	// 校验本地缓存中的数据
	wantItems1 := make([]domain.RankItem, 0, 100)
	for i := 1; i <= 100; i++ {
		wantItems1 = append(wantItems1, domain.RankItem{
			ID:    int64(i),
			Score: 4000 - i,
		})
	}
	t.checkItems(items, wantItems1)
	// 更新数据
	err = t.svc.Update(context.Background(), []domain.RankItem{
		{
			ID:    99,
			Score: 9999,
		},
	})
	require.NoError(t.T(), err)
	t.updateLocalCache()
	wantItems2 := t.getWantItems2()
	items, err = t.svc.TopN(context.Background())
	t.checkItems(items, wantItems2)
	// 模拟更新第三方服务的数据
	t.updateTriSvc()
	// 模拟定时任务，更新redis
	t.updateRedis()
	// 校验redis中的数据是否已经同步
	t.checkCacheData()
	// 模拟定时任务，更新本地缓存
	t.updateLocalCache()
	// 获取本地缓存中的数据
	items, err = t.svc.TopN(context.Background())
	require.NoError(t.T(), err)
	// 校验本地缓存中的数据
	wantItems3 := t.getWantItems3()
	t.checkItems(items, wantItems3)
}

func (t *TestSuite) getWantItems3() []domain.RankItem {
	wantItems3 := make([]domain.RankItem, 0, 100)
	for i := 4999; i >= 4900; i-- {
		wantItems3 = append(wantItems3, domain.RankItem{
			ID:    int64(i),
			Score: i,
		})
	}
	return wantItems3
}

func (t *TestSuite) getWantItems2() []domain.RankItem {
	wantItems2 := []domain.RankItem{
		{
			ID:    99,
			Score: 9999,
		},
	}
	for i := 1; i <= 100; i++ {
		if i == 99 {
			continue
		}
		wantItems2 = append(wantItems2, domain.RankItem{
			ID:    int64(i),
			Score: 4000 - i,
		})
	}
	return wantItems2
}

func (t *TestSuite) checkCacheData() {
	wantItems := make([]domain.RankItem, 0, 1000)
	for i := 4999; i >= 4000; i-- {
		wantItems = append(wantItems, domain.RankItem{
			ID:    int64(i),
			Score: i,
		})
	}
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
	t.localCache = localCache
	t.redisCache = redisCache

	t.triSvc = t.initTriSvc()
	// 定时任务
	//err := cronjob.InitJob(t.triSvc, redisCache, localCache)
	//require.NoError(t.T(), err)
	repo := repository.NewRankRepository(localCache, redisCache)
	t.svc = service.NewRankService(repo)
	// 更新redis和本地缓存
	t.updateRedis()
	t.updateLocalCache()
}

func (t *TestSuite) initTriSvc() cronjob.TriSvc {
	db := test.InitDB()
	module, err := tri.InitModule(db)
	require.NoError(t.T(), err)
	t.articleSvc = module.Svc
	// 初始化文章服务的数据
	articles := make([]triDomain.Article, 0, 4000)
	for i := 1; i <= 4000; i++ {
		articles = append(articles, triDomain.Article{
			ID:      int64(i),
			LikeCnt: int32(4000 - i),
		})
	}
	err = t.articleSvc.BatchCreate(context.Background(), articles)
	require.NoError(t.T(), err)
	triService := cronjob.NewTriSvc(t.articleSvc)
	return triService
}

func (t *TestSuite) updateLocalCache() {
	cronjob.NewRedisToLocalJob(t.redisCache, t.localCache, 100).Run()
}

func (t *TestSuite) updateRedis() {
	cronjob.NewDBToRedisJob(t.redisCache, t.triSvc, 1000).Run()
}

func (t *TestSuite) updateTriSvc() {
	articles := make([]triDomain.Article, 0, 4000)
	for i := 4000; i < 5000; i++ {
		articles = append(articles, triDomain.Article{
			ID:      int64(i),
			LikeCnt: int32(i),
		})
	}
	err := t.articleSvc.BatchCreate(context.Background(), articles)
	require.NoError(t.T(), err)
}

func TestTopN(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

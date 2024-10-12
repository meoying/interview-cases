package case21

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"interview-cases/case21_30/case21/cronjob"
	"interview-cases/case21_30/case21/cronjob/dao"
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
	activitySvc cronjob.ArticleSvc
	svc         service.RankService
	localCache  *local.Cache
	redisCache  *redis.Cache
}

func (t *TestSuite) SetupSuite() {
	// 初始化
	t.initServer()
}

// 运行流程
func (t *TestSuite) TestRank() {
	// 初始化成功，查看榜单数据
	t.testRankNormal()

	// 某个文章数更新，榜单排名随之变化
	t.testRankUpdate()

	// 测试文章服务排名数据更新了，同步到redis
	t.testRankSync()

}

func (t *TestSuite) testRankUpdate() {
	// 文章id为99的点赞数暴涨到9999
	// 模拟文章服务更新
	log.Println("执行TestRankUpdate更新数据。。。")
	t.activitySvc.BatchCreate(context.Background(), []domain.Article{
		{
			ID:      99,
			LikeCnt: 9999,
		},
	})
	// 榜单数据更新
	err := t.svc.Update(context.Background(), []domain.Article{
		{
			ID:      99,
			LikeCnt: 9999,
		},
	})
	require.NoError(t.T(), err)
	log.Println("TestRankUpdate更新数据成功。。。")
	log.Println("TestRankUpdate等待一分钟同步到本地缓存。。。")
	// 等待一分钟同步到本地缓存
	time.Sleep(1*time.Minute + 5*time.Second)
	wantItems2 := t.getWantItems2()
	// 获取榜单数据
	items, err := t.svc.TopN(context.Background())
	// 校验榜单数据
	t.checkItems(items, wantItems2)
}

func (t *TestSuite) testRankNormal() {
	// 查看前100名
	items, err := t.svc.TopN(context.Background())
	require.NoError(t.T(), err)
	// 校验本地缓存中的数据
	wantItems1 := make([]domain.Article, 0, 100)
	for i := 1; i <= 100; i++ {
		wantItems1 = append(wantItems1, domain.Article{
			ID:      int64(i),
			LikeCnt: 4000 - i,
		})
	}
	t.checkItems(items, wantItems1)
}

func (t *TestSuite) testRankSync() {
	// 模拟第三方服务数据更新
	t.updateTriSvc()
	log.Println("执行TestRankSync更新数据成功。。。")
	// 等待2分钟同步到redis
	log.Println("TestRankUpdate等待两分钟同步到redis。。。")
	time.Sleep(2*time.Minute + 10*time.Second)
	// 校验redis中的数据是否已经同步
	t.checkCacheData()
	// 等待1分钟同步到本地缓存
	log.Println("TestRankUpdate等待一分钟同步到本地缓存。。。")
	time.Sleep(1*time.Minute + 10*time.Second)
	// 获取本地缓存中的数据
	items, err := t.svc.TopN(context.Background())
	require.NoError(t.T(), err)
	// 校验本地缓存中的数据
	wantItems3 := t.getWantItems3()
	t.checkItems(items, wantItems3)
}

func (t *TestSuite) getWantItems3() []domain.Article {
	wantItems3 := []domain.Article{
		{
			ID:      99,
			LikeCnt: 9999,
		},
	}
	for i := 4999; i >= 4901; i-- {
		wantItems3 = append(wantItems3, domain.Article{
			ID:      int64(i),
			LikeCnt: i,
		})
	}
	return wantItems3
}

func (t *TestSuite) getWantItems2() []domain.Article {
	wantItems2 := []domain.Article{
		{
			ID:      99,
			LikeCnt: 9999,
		},
	}
	for i := 1; i <= 100; i++ {
		if i == 99 {
			continue
		}
		wantItems2 = append(wantItems2, domain.Article{
			ID:      int64(i),
			LikeCnt: 4000 - i,
		})
	}
	return wantItems2
}

func (t *TestSuite) checkCacheData() {
	wantItems := []domain.Article{
		{
			ID:      99,
			LikeCnt: 9999,
		},
	}
	for i := 4999; i >= 4001; i-- {
		wantItems = append(wantItems, domain.Article{
			ID:      int64(i),
			LikeCnt: i,
		})
	}
	actualItems, err := t.redisCache.Get(context.Background(), 1000)
	require.NoError(t.T(), err)
	assert.Equal(t.T(), wantItems, actualItems)
}

// 校验获取数据是否正确
func (t *TestSuite) checkItems(actualItems []domain.Article, wantItems []domain.Article) {
	assert.Equal(t.T(), wantItems, actualItems)
}

func (t *TestSuite) initServer() {
	client := test.InitRedis()
	localCache := local.NewCache()
	redisCache := redis.NewCache(client, "rank")
	t.localCache = localCache
	t.redisCache = redisCache
	// 初始化文章服务
	t.initArticleSvc()
	// 定时任务
	err := cronjob.InitJob(t.activitySvc, redisCache, localCache)
	require.NoError(t.T(), err)
	repo := repository.NewRankRepository(localCache, redisCache)
	t.svc = service.NewRankService(repo)
	// 初始化redis和本地缓存中的数据
	t.updateRedis()
	t.updateLocalCache()
}

func (t *TestSuite) initArticleSvc() {
	db := test.InitDB()
	articleDao := dao.NewArticleStaticDAO(db)
	articleSvc := cronjob.NewArticleSvc(articleDao)
	t.activitySvc = articleSvc
	// 初始化文章服务的数据
	articles := make([]domain.Article, 0, 4000)
	for i := 1; i <= 4000; i++ {
		articles = append(articles, domain.Article{
			ID:      int64(i),
			LikeCnt: 4000 - i,
		})
	}
	err := t.activitySvc.BatchCreate(context.Background(), articles)
	require.NoError(t.T(), err)
	return
}

func (t *TestSuite) updateLocalCache() {
	cronjob.NewRedisToLocalJob(t.redisCache, t.localCache, 100).Run()
}

func (t *TestSuite) updateRedis() {
	cronjob.NewDBToRedisJob(t.redisCache, t.activitySvc, 1000).Run()
}

func (t *TestSuite) updateTriSvc() {
	articles := make([]domain.Article, 0, 4000)
	for i := 4000; i < 5000; i++ {
		articles = append(articles, domain.Article{
			ID:      int64(i),
			LikeCnt: i,
		})
	}
	err := t.activitySvc.BatchCreate(context.Background(), articles)
	require.NoError(t.T(), err)
}

// 注意运行前删除dockerCompose上所有容器，再新建
func TestTopN(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

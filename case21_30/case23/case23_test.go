package case23

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"interview-cases/case21_30/case23/cronjob"
	"interview-cases/case21_30/case23/domain"
	"interview-cases/case21_30/case23/repository"
	"interview-cases/case21_30/case23/repository/cache/local"
	"interview-cases/case21_30/case23/repository/cache/redis"
	"interview-cases/case21_30/case23/service"
	"interview-cases/test"
	"log"
	"sort"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	svc service.RankService
}

func (t *TestSuite) SetupSuite() {
	client := test.InitRedis()
	redisCache := redis.NewCache(client)
	localCache := local.NewCache()
	rankRepo := repository.NewRankRepo(localCache, redisCache)
	rankSvc := service.NewRankSvc(rankRepo)
	t.svc = rankSvc
	err := cronjob.InitJob(redisCache, localCache)
	require.NoError(t.T(), err)
	// 往排行榜初始化100000个元素
	for i := 1; i <= 100000; i++ {
		err = rankSvc.UpdateScore(context.Background(), domain.RankItem{
			ID:    int64(i),
			Score: int64(i),
		})
		require.NoError(t.T(), err)
	}
	log.Println("初始化成功")
	time.Sleep(1 * time.Minute)
}

func (t *TestSuite) TestRank() {
	wantItems := make([]domain.RankItem, 0, 100)
	for i := 100000; i >= 99901; i-- {
		wantItems = append(wantItems, domain.RankItem{
			ID:    int64(i),
			Score: int64(i),
		})
	}
	actualItems, err := t.svc.TopN(context.Background())
	require.NoError(t.T(), err)
	assert.Equal(t.T(), wantItems, actualItems)
	// 更新排行榜的元素
	wantItems = t.updateRankItems()
	// 睡眠一分钟等待数据同步到本地缓存
	time.Sleep(1 * time.Minute)
	actualItems, err = t.svc.TopN(context.Background())
	require.NoError(t.T(), err)
	assert.Equal(t.T(), actualItems, wantItems)
}

func (t *TestSuite) updateRankItems() []domain.RankItem {
	items := make([]domain.RankItem, 0, 100)
	for i := 200000; i < 200050; i++ {
		item := domain.RankItem{
			ID:    int64(i),
			Score: int64(i),
		}
		err := t.svc.UpdateScore(context.Background(), item)
		require.NoError(t.T(), err)
		items = append(items, item)
	}
	for i := 100000; i > 99950; i-- {
		item := domain.RankItem{
			ID:    int64(i),
			Score: int64(i) + 500,
		}
		err := t.svc.UpdateScore(context.Background(), item)
		require.NoError(t.T(), err)
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})
	return items

}

func TestTopN(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

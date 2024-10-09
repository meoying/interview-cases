package cronjob

import (
	"github.com/robfig/cron/v3"
	"interview-cases/case21_30/case21/repository/cache/local"
	"interview-cases/case21_30/case21/repository/cache/redis"
)

func InitJob(triSvc TriSvc, redisCache *redis.Cache, localCache *local.Cache) error {
	c := cron.New()
	// 这边为了测试方便，每三分钟将全局排行前1000的数据同步进redis，真正回答的时候可以说大概一个小时
	_, err := c.AddJob("*/3 * * * *", NewDBToRedisJob(redisCache, triSvc,1000))
	if err != nil {
		return err
	}
	// 每分钟，将redis中排行前100的数据同步到本地缓存
	_, err = c.AddJob("*/1 * * * *", NewRedisToLocalJob(redisCache, localCache,100))
	if err != nil {
		return err
	}
	c.Start()
	return nil
}

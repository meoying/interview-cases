package cronjob

import (
	"github.com/robfig/cron/v3"
	"interview-cases/case21_30/case21/repository/cache/local"
	"interview-cases/case21_30/case21/repository/cache/redis"
)

func InitJob(triSvc TriSvc, redisCache *redis.Cache, localCache *local.Cache) error {
	c := cron.New()
	// 一个小时执行一次
	_, err := c.AddJob("1 * * * *", NewDBToRedisJob(redisCache, triSvc, 1000))
	if err != nil {
		return err
	}
	// 每分钟，将redis中排行前100的数据同步到本地缓存
	_, err = c.AddJob("*/1 * * * *", NewRedisToLocalJob(redisCache, localCache, 100))
	if err != nil {
		return err
	}
	c.Start()
	return nil
}

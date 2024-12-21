package cronjob

import (
	"github.com/robfig/cron/v3"
	"interview-cases/case21_30/case21/repository/cache/local"
	"interview-cases/case21_30/case21/repository/cache/redis"
)

func InitJob(articleSvc ArticleSvc, redisCache *redis.Cache, localCache *local.Cache) error {
	c := cron.New()
	// 从全局同步不需要太频繁大概一小时一次就可以,这边为了运行效果改成了每两分钟一次
	_, err := c.AddJob("*/2 * * * *", NewDBToRedisJob(redisCache, articleSvc, 1000))
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

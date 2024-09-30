package cronjob

import (
	"github.com/robfig/cron/v3"
	"interview-cases/case21_30/case21/repository/cache/local"
	"interview-cases/case21_30/case21/repository/cache/redis"
)

func InitJob(triSvc TriSvc, redisCache *redis.Cache, localCache *local.Cache) error {
	c := cron.New()
	// 每两分钟将db同步进redis，真正生产可以修改成1小时一次
	_, err := c.AddJob("*/2 * * * *", NewDBToRedisJob(redisCache, triSvc))
	if err != nil {
		return err
	}
	// 每分钟，将redis中的数据同步到本地缓存
	_, err = c.AddJob("*/1 * * * *", NewRedisToLocalJob(redisCache, localCache))
	if err != nil {
		return err
	}
	c.Start()
	return nil
}

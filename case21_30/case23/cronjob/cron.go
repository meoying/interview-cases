package cronjob

import (
	"github.com/robfig/cron/v3"
	"interview-cases/case21_30/case23/repository/cache/local"
	"interview-cases/case21_30/case23/repository/cache/redis"
)

func InitJob(redisCache *redis.Cache, localCache *local.Cache) error {
	c := cron.New()
	// 每分钟，将redis中的数据同步到本地缓存
	_, err := c.AddJob("*/1 * * * *", NewRedisToLocalJob(redisCache, localCache))
	if err != nil {
		return err
	}
	c.Start()
	return nil
}

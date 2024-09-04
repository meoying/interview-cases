package repo

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	ratelimiter "interview-cases/case11/ratelimit"
)

type Article struct {
	ID      int32
	Title   string
	Author  string
	Content string
}

var (
	client      redis.Cmdable
	db          *gorm.DB
	tokenBucket *ratelimiter.TokenBucket
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/interview_cases"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&Article{})
	if err != nil {
		panic(err)
	}
	return db

}

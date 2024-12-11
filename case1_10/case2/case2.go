package case2

import (
	"context"
	"gorm.io/gorm"
	"math/rand/v2"
	"time"
)

type Order struct {
	ID         int   `gorm:"primaryKey;autoIncrement"` // ID
	Uid        int64 `gorm:"type:int;index:create_time_uid_idx"`
	Credit     int
	CreateTime time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index:create_time_uid_idx"` // 创建时间
}

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&Order{})
}

func InitData(db *gorm.DB) error {
	baseCreated, err := time.Parse(time.DateOnly, "2024-11-01")
	if err != nil {
		return err
	}
	// 大商家123
	for i := 0; i < 10000; i++ {
		err := db.WithContext(context.Background()).Create(&Order{
			Uid:        123,
			Credit:     i,
			CreateTime: baseCreated.Add(-1 * time.Duration(rand.IntN(1000)) * time.Hour),
		}).Error
		if err != nil {
			continue
		}
	}
	// 小商家234
	for i := 0; i < 1000; i++ {
		err := db.WithContext(context.Background()).Create(&Order{
			Uid:        234,
			Credit:     i,
			CreateTime: baseCreated.Add(-1 * time.Duration(rand.IntN(1000)) * time.Hour),
		}).Error
		if err != nil {
			continue
		}
	}
	// 小商家456
	for i := 0; i < 1000; i++ {
		err := db.WithContext(context.Background()).Create(&Order{
			Uid:        456,
			Credit:     i,
			CreateTime: baseCreated.Add(-1 * time.Duration(rand.IntN(1000)) * time.Hour),
		}).Error
		if err != nil {
			continue
		}
	}

	return nil
}

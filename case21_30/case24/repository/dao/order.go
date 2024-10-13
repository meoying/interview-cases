package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Order struct {
	ID      int64 `gorm:"primaryKey,autoIncrement"`
	Name    string
	Price   int32
	BuyerID int64
	Ctime   int64
	Utime   int64
}

type OrderDAO interface {
	Save(ctx context.Context, order Order) error
	Get(ctx context.Context, id int64) (Order, error)
}

type orderDao struct {
	db *gorm.DB
}

func NewOrderDAO(db *gorm.DB) OrderDAO {
	return &orderDao{
		db: db,
	}
}

func (o *orderDao) Save(ctx context.Context, order Order) error {
	now := time.Now().UnixMilli()
	order.Ctime = now
	order.Utime = now
	return o.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{
				Name: "id",
			},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"price",
			"utime",
		}),
	}).Create(order).Error
}

func (o *orderDao) Get(ctx context.Context, id int64) (Order, error) {
	var order Order
	// 使用 GORM 的 First 方法按 ID 查询订单
	err := o.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	return order, err
}

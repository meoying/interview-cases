package case28

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Order struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"` // 订单ID，主键
	UserID    int64 `gorm:"index"`                    // 用户ID，建立索引
	ProductID int64 // 商品ID
	// 为了方便我就不建商品表了，直接冗余进order表
	ProductName string // 商品名
	Quantity    int    // 商品数量
	Status      int32  // 订单状态 0-未支付 1-已支付
	Ctime       int64  // 订单创建时间
	Utime       int64  // 订单更新时间
}

type OrderDAO interface {
	BatchCreate(ctx context.Context, order []Order) error
	// UpdateStatus 更新订单状态
	UpdateStatus(ctx context.Context, id int64, status int32) error
	// 批量查找订单
	BatchGetOrder(ctx context.Context, ids []int64) (map[int64]*Order, error)
	// 单个获取订单
	GetOrder(ctx context.Context, id int64) (Order, error)
}

// OrderDAOImpl 是 OrderDAO 接口的实现
type OrderDAOImpl struct {
	db *gorm.DB
}

// NewOrderDAO 创建 OrderDAO 实例
func NewOrderDAO(db *gorm.DB) OrderDAO {
	return &OrderDAOImpl{db: db}
}

func (dao *OrderDAOImpl) GetOrder(ctx context.Context, id int64) (Order, error) {
	var order Order
	err := dao.db.WithContext(ctx).
		Model(&Order{}).
		Where("id = ?", id).
		First(&order).Error
	return order, err
}

// Create 插入一条新的订单记录
func (dao *OrderDAOImpl) BatchCreate(ctx context.Context, orders []Order) error {
	if len(orders) == 0 {
		return fmt.Errorf("订单不能为空")
	}
	for idx, order := range orders {
		order.Ctime = time.Now().Unix()
		order.Utime = time.Now().Unix()
		orders[idx] = order
	}
	// 插入订单记录
	return dao.db.WithContext(ctx).Create(orders).Error
}

// UpdateStatus 更新订单的状态
func (dao *OrderDAOImpl) UpdateStatus(ctx context.Context, id int64, status int32) error {
	// 更新状态字段
	return dao.db.WithContext(ctx).Model(&Order{}).Where("id = ?", id).Update("status", status).Error
}

// BatchGetOrder 根据订单ID批量查询订单
func (dao *OrderDAOImpl) BatchGetOrder(ctx context.Context, ids []int64) (map[int64]*Order, error) {
	var orders []Order
	// 执行批量查询
	err := dao.db.WithContext(ctx).Where("id IN ?", ids).Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("批量查询订单失败: %w", err)
	}
	// 将查询结果转换为 map，键为订单ID
	orderMap := make(map[int64]*Order, len(orders))
	for _, order := range orders {
		orderMap[order.ID] = &order
	}
	return orderMap, nil
}
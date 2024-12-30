package case28

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// Notification 定义通知的结构体
type Notification struct {
	ID      int64  `gorm:"primaryKey;autoIncrement"` // 通知ID，主键
	UserID  int64  `gorm:"index"`                    // 接收通知的用户ID
	Biz     string `gorm:"index:idx_biz_bizid"`
	BizID   int64  `gorm:"index:idx_biz_bizid"` // 相关订单ID (如果通知与订单相关)
	Message string `gorm:"type:text"`           // 通知的内容或消息
	Type    string `gorm:"type:varchar(50)"`    // 通知类型
	Status  int32  // 通知状态 0-未读 1-已读
	Ctime   int64  // 订单创建时间
	Utime   int64  // 订单更新时间
}

type NotificationDAO interface {
	Create(ctx context.Context, notification *Notification) error
	BatchCreate(ctx context.Context, notifications []*Notification) error
}

// NotificationDAOImpl 是 NotificationDAO 的具体实现
type NotificationDAOImpl struct {
	db *gorm.DB
}

// NewNotificationDAO 创建 NotificationDAO 实例
func NewNotificationDAO(db *gorm.DB) NotificationDAO {
	return &NotificationDAOImpl{db: db}
}

// Create 插入一条通知记录
func (dao *NotificationDAOImpl) Create(ctx context.Context, notification *Notification) error {
	if notification == nil {
		return fmt.Errorf("通知对象不能为空")
	}
	notification.Ctime = time.Now().Unix()
	notification.Utime = time.Now().Unix()

	return dao.db.WithContext(ctx).Create(notification).Error
}

// BatchCreate 批量插入通知记录
func (dao *NotificationDAOImpl) BatchCreate(ctx context.Context, notifications []*Notification) error {
	if len(notifications) == 0 {
		return fmt.Errorf("通知列表不能为空")
	}
	for idx, notification := range notifications {
		notification.Ctime = time.Now().Unix()
		notification.Utime = time.Now().Unix()
		notifications[idx] = notification
	}
	return dao.db.WithContext(ctx).Create(&notifications).Error
}

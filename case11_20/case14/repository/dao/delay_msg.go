package dao

import (
	"context"
	"gorm.io/gorm"
	"interview-cases/case11_20/case14/domain"
	"time"
)

type DelayMsgDAO interface {
	// 批量添加
	Insert(ctx context.Context, msg []DelayMsg) error
	// 批量更新成完成
	Complete(ctx context.Context, id []int64) error
	FindDelayMsg(ctx context.Context) ([]DelayMsg, error)
}

type DelayMsg struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Topic    string `gorm:"type=varchar(512)"`
	Value    string `gorm:"text"`
	Deadline int64
	Status   uint8 `gorm:"type:tinyint(3);comment:0-待完成 1-完成"`
	Ctime    int64
	Utime    int64 `gorm:"index"`
}

type delayMsgDAO struct {
	db *gorm.DB
}

func (d *delayMsgDAO) Insert(ctx context.Context, msg []DelayMsg) error {
	now := time.Now().UnixMilli()
	for i := range msg {
		msg[i].Ctime = now
		msg[i].Utime = now
	}
	return d.db.WithContext(ctx).Create(&msg).Error
}

func (d *delayMsgDAO) Complete(ctx context.Context, id []int64) error {
	return d.db.WithContext(ctx).Where("id in (?)", id).Updates(map[string]any{
		"status": 1,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (d *delayMsgDAO) FindDelayMsg(ctx context.Context) ([]DelayMsg, error) {
	// 找到到时间的延迟消息
	var msgs []DelayMsg
	err := d.db.WithContext(ctx).
		Where("deadline > ? and status = ?",time.Now().UnixMilli(),domain.Waiting.ToUint8()).
		Scan(&msgs).Error
	return msgs, err
}

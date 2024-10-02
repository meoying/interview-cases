package dao

import (
	"context"
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"sync/atomic"
	"time"
)

type DelayMsg struct {
	Id    int64  `gorm:"primaryKey"`
	Topic string `gorm:"type=varchar(512)"`
	Value []byte `gorm:"type=BLOB"`
	// 创建一个唯一索引，这个列可以为空
	// 你也可以从业务层面上强制要求它们不为空
	Key      sql.NullString `gorm:"unique;type:varchar(512)"`
	Deadline int64          `grom:"index"`
	Status   uint8          `gorm:"type:tinyint(3);comment:0-待完成 1-完成"`
	Ctime    int64
	Utime    int64 `gorm:"index"`
}

type DelayMsgDAO struct {
	db     *gorm.DB
	tables []string
	index  atomic.Int64
}

// NewDelayMsgDAO 如果你有多个集群，那么这里传入多个 db 来轮询
func NewDelayMsgDAO(db *gorm.DB) *DelayMsgDAO {
	return &DelayMsgDAO{
		db: db,
		tables: []string{
			"delay_msg_db_0.delay_msg_tab_0",
			"delay_msg_db_0.delay_msg_tab_1",
			"delay_msg_db_1.delay_msg_tab_0",
			"delay_msg_db_1.delay_msg_tab_1",
		},
	}
}

func (d *DelayMsgDAO) getTables() []string {
	tabNames := make([]string, 0)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			tabNames = append(tabNames, fmt.Sprintf("delay_msg_db_%d.delay_msg_tab_%d", i, j))
		}
	}
	return tabNames
}

// Insert 轮询加入
func (d *DelayMsgDAO) Insert(ctx context.Context, msg DelayMsg) error {
	now := time.Now().UnixMilli()
	msg.Ctime = now
	msg.Utime = now
	// 等待被转发
	msg.Status = 0
	// 轮询
	idx := d.index.Add(1) % int64(len(d.tables))
	// 要指定表
	return d.db.WithContext(ctx).Table(d.tables[idx]).Create(&msg).Error
}

func (d *DelayMsgDAO) Complete(ctx context.Context, tab string, ids ...int64) error {
	err := d.db.WithContext(ctx).Table(tab).Where("id in (?)", ids).Updates(map[string]any{
		"status": 1,
		"utime":  time.Now().UnixMilli(),
	}).Error
	return err
}

// FindDelayMsg 广播每次每个库最多拿10个
func (d *DelayMsgDAO) FindDelayMsg(ctx context.Context, tab string, limit int) ([]DelayMsg, error) {
	// 找到到时间的延迟消息,每个库拿10个
	var ms []DelayMsg
	err := d.db.WithContext(ctx).Table(tab).
		Where("status = ? and deadline  <= ?", 0, time.Now().UnixMilli()).
		Order("ctime desc").
		Limit(limit).
		Find(&ms).Error
	return ms, err
}

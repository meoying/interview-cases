package dao

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultGetMsgLimit = 10
)

type DelayMsgDAO interface {
	// 批量添加
	Insert(ctx context.Context, msg DelayMsg) error
	// 批量更新成完成
	Complete(ctx context.Context, id []int64) error
	FindDelayMsg(ctx context.Context) ([]DelayMsg, error)
}

type DelayMsg struct {
	Id       int64  `gorm:"primaryKey"`
	Topic    string `gorm:"type=varchar(512)"`
	Value    string `gorm:"text"`
	Deadline int64  `grom:"index"`
	Status   uint8  `gorm:"type:tinyint(3);comment:0-待完成 1-完成"`
	Ctime    int64
	Utime    int64 `gorm:"index"`
}

type delayMsgDAO struct {
	db *gorm.DB
	// 使用雪花id保证id不重复
	node  *snowflake.Node
	count int64
}

func NewDelayMsgDAO(db *gorm.DB) (DelayMsgDAO, error) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, err
	}

	return &delayMsgDAO{
		db:    db,
		count: 0,
		node:  node,
	}, nil
}

// 假设库有两个 每个库有两个表 供轮询插入使用
func (d *delayMsgDAO) getTable() string {
	count := atomic.AddInt64(&d.count, 1)
	dbName := fmt.Sprintf("delay_msg_db_%d", (count%4)/2)
	tabName := fmt.Sprintf("delay_msg_tab_%d", (count%4)%2)
	return fmt.Sprintf("%s.%s", dbName, tabName)
}

func (d *delayMsgDAO) getTables() []string {
	tabNames := make([]string, 0)
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			tabNames = append(tabNames, fmt.Sprintf("delay_msg_db_%d.delay_msg_tab_%d", i, j))
		}
	}
	return tabNames
}

// 轮询加入
func (d *delayMsgDAO) Insert(ctx context.Context, msg DelayMsg) error {
	now := time.Now().UnixMilli()
	msg.Ctime = now
	msg.Utime = now
	msg.Id = d.node.Generate().Int64()
	return d.db.WithContext(ctx).Table(d.getTable()).Create(&msg).Error
}

func (d *delayMsgDAO) Complete(ctx context.Context, id []int64) error {
	tabs := d.getTables()
	var eg errgroup.Group
	fmt.Println("")
	for _, tab := range tabs {
		tab := tab
		eg.Go(func() error {
			return d.db.WithContext(ctx).Table(tab).Where("id in (?)", id).Updates(map[string]any{
				"status": 1,
				"utime":  time.Now().UnixMilli(),
			}).Error
		})
	}
	return eg.Wait()

}

// 广播每次每个库最多拿10个
func (d *delayMsgDAO) FindDelayMsg(ctx context.Context) ([]DelayMsg, error) {
	// 找到到时间的延迟消息,每个库拿10个
	var (
		eg   errgroup.Group
		msgs []DelayMsg
	)
	mu := &sync.RWMutex{}
	tabs := d.getTables()
	for _, tabName := range tabs {
		tabName := tabName
		eg.Go(func() error {
			var ms []DelayMsg
			d.db.WithContext(ctx).Table(tabName).
				Where("status = ? and deadline  < ?", 0, time.Now().UnixMilli()).
				Order("ctime desc").
				Limit(defaultGetMsgLimit).
				Find(&ms)
			mu.Lock()
			msgs = append(msgs, ms...)
			mu.Unlock()
			return nil
		})
	}
	return msgs, eg.Wait()
}

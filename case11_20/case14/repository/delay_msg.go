package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"interview-cases/case11_20/case14/domain"
	"interview-cases/case11_20/case14/repository/dao"
	"time"
)

type DelayMsgRepository interface {
	Insert(ctx context.Context, msg domain.DelayMsg) error
	// 批量更新成完成
	Complete(ctx context.Context, ids []int64) error
	FindDelayMsg(ctx context.Context) ([]domain.DelayMsg, error)
}

type DelayMsgRepo struct {
	delayDao dao.DelayMsgDAO
}

func NewDelayMsgRepo(delayDao dao.DelayMsgDAO) DelayMsgRepository {
	return &DelayMsgRepo{delayDao: delayDao}
}

func (d *DelayMsgRepo) Insert(ctx context.Context, msg domain.DelayMsg) error {
	return d.delayDao.Insert(ctx,d.toEntity(msg))
}

func (d *DelayMsgRepo) Complete(ctx context.Context, ids []int64) error {
	return d.delayDao.Complete(ctx, ids)
}

func (d *DelayMsgRepo) FindDelayMsg(ctx context.Context) ([]domain.DelayMsg, error) {
	msgs, err := d.delayDao.FindDelayMsg(ctx)
	if err != nil {
		return nil, err
	}
	return slice.Map(msgs, func(idx int, src dao.DelayMsg) domain.DelayMsg {
		return d.toDomain(src)
	}),nil
}

func (d *DelayMsgRepo) toEntity(msg domain.DelayMsg) dao.DelayMsg {
	return dao.DelayMsg{
		Id:       msg.ID,
		Topic:    msg.Topic,
		Value:    msg.Value,
		Deadline: msg.Deadline.UnixMilli(),
	}
}

func (d *DelayMsgRepo) toDomain(msg dao.DelayMsg) domain.DelayMsg {
	return domain.DelayMsg{
		ID:       msg.Id,
		Topic:    msg.Topic,
		Value:    msg.Value,
		Deadline: time.UnixMilli(msg.Deadline),
		Status:   domain.DelayMsgStatus(msg.Status),
		Ctime:    time.UnixMilli(msg.Ctime),
		Utime:    time.UnixMilli(msg.Utime),
	}
}

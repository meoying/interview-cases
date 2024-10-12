package delay_platform

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ecodeclub/ekit/sqlx"
	"interview-cases/case11_20/case15/delay_platform/dao"
	"log/slog"
	"time"
)

// DelayMsgReceiver 延迟消息接收者
type DelayMsgReceiver struct {
	consumer *kafka.Consumer
	dao      *dao.DelayMsgDAO
}

func NewDelayMsgReceiver(consumer *kafka.Consumer, dao *dao.DelayMsgDAO) *DelayMsgReceiver {
	return &DelayMsgReceiver{consumer: consumer, dao: dao}
}

// ReceiveMsg 接收消息进行转发
func (receiver *DelayMsgReceiver) ReceiveMsg() {
	for {
		msg, err := receiver.consumer.ReadMessage(-1)
		if err != nil {
			// 失败记录一下报错然后重试
			// 这里如果一直失败其实也没什么很好的办法，可以告警，然后人手工介入处理
			slog.Error("获取延迟消息失败", slog.Any("err", err))
			continue
		}
		vv, _ := json.Marshal(msg)
		slog.Info("成功获取延迟消息", slog.Any("msg", string(vv)))
		err = receiver.sendToDb(msg)
		if err != nil {
			slog.Error("转储延迟消息失败", slog.Any("err", err))
		} else {
			slog.Info("成功转储延迟消息", slog.Any("msg", string(vv)))
		}
	}
}

func (receiver *DelayMsgReceiver) sendToDb(msg *kafka.Message) error {
	type DelayMsg struct {
		// 转发内容
		Value []byte
		// 转发主题，也就是业务主题
		Topic string
		// 到什么时候发出去
		Deadline int64
		Key      string
	}
	var delayMsg DelayMsg
	err := json.Unmarshal(msg.Value, &delayMsg)
	if err != nil {
		return fmt.Errorf("序列化延迟消息失败  %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	now := time.Now().UnixMilli()
	err = receiver.dao.Insert(ctx, dao.DelayMsg{
		Topic:    delayMsg.Topic,
		Value:    delayMsg.Value,
		Deadline: delayMsg.Deadline,
		Key:      sqlx.NewNullString(delayMsg.Key),
		Status:   DelayMsgStatusWaiting.ToUint8(),
		Ctime:    now,
		Utime:    now,
	})
	if err != nil {
		return fmt.Errorf("转储消息失败 %w", err)
	}
	_, err = receiver.consumer.CommitMessage(msg)
	if err != nil {
		return fmt.Errorf("转储消息失败 %w", err)
	}
	return nil
}

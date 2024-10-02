package delay_platform

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ecodeclub/ekit/syncx"
	"interview-cases/case11_20/case15/delay_platform/dao"
	"log/slog"
	"sync"
	"time"
)

// DelayMsgSender 延迟消息发送者
type DelayMsgSender struct {
	topicConn *syncx.Map[string, *kafka.Producer]
	dao       *dao.DelayMsgDAO
	// 轮询的目标表
	dst string
}

func NewDelayMsgSender(topicConn *syncx.Map[string, *kafka.Producer],
	dao *dao.DelayMsgDAO,
	dst string,
) *DelayMsgSender {
	return &DelayMsgSender{topicConn: topicConn, dao: dao, dst: dst}
}

func (sender *DelayMsgSender) SendMsg() {
	// 每30s轮询一次，根据自己的业务进行修改
	for {
		cnt := sender.oneLoop()
		if cnt == 0 {
			// 说明暂时没有要发送的消息，所以可以睡眠 1s，因为我们允许误差在秒级
			// 而后继续下一个循环
			time.Sleep(time.Second)
		}
	}
}

func (sender *DelayMsgSender) oneLoop() int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return sender.sendMsgs(ctx)
}

func (sender *DelayMsgSender) sendMsgs(ctx context.Context) int {
	// 获取需要发送的消息
	// 10 个一批
	msgs, err := sender.dao.FindDelayMsg(ctx, sender.dst, 10)
	if err != nil {
		slog.Error("获取延迟消息失败", slog.Any("err", err))
	}

	if len(msgs) == 0 {
		return 0
	}
	// 准备发送
	var wg sync.WaitGroup
	for _, msg := range msgs {
		wg.Add(1)
		go func(msg dao.DelayMsg) {
			defer wg.Done()
			newctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			eerr := sender.sendMsg(newctx, msg)
			if eerr != nil {
				slog.Error("转发延迟消息失败", slog.Any("err", err))
			} else {
				slog.Info("成功转发延迟消息", slog.Any("msg", msg.Value))
			}
		}(msg)
	}
	wg.Wait()
	return len(msgs)
}

func (sender *DelayMsgSender) sendMsg(ctx context.Context, msg dao.DelayMsg) error {
	conn, ok := sender.topicConn.Load(msg.Topic)
	if !ok {
		return fmt.Errorf("未知topic %s", msg.Topic)
	}
	topic := msg.Topic
	conn.BeginTransaction()
	err := conn.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic},
		Value:          msg.Value,
		Key:            []byte(msg.Key.String),
	}, nil)
	if err != nil {
		return fmt.Errorf("消息 %v 转发失败 %w", msg, err)
	}
	err = sender.dao.Complete(ctx, sender.dst, msg.Id)
	if err != nil {
		return fmt.Errorf("更新消息%v 失败 %w", msg, err)
	}
	return nil
}

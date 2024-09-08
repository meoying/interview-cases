package case14

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ecodeclub/ekit/syncx"
	"interview-cases/case11_20/case14/domain"
	"interview-cases/case11_20/case14/service"
	"log/slog"
	"sync"
	"time"
)

const delayTopic = "delay_topic"

// 延迟消息接收者
type DelayMsgReceiver struct {
	consumer *kafka.Consumer
	svc      service.DelayMsgSvc
}

// 接收消息进行转发
func (receiver *DelayMsgReceiver) ReceiveMsg() {
	for {
		msg, err := receiver.consumer.ReadMessage(-1)
		if err != nil {
			// 失败记录一下报错然后重试
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
	var delayMsg domain.DelayMsg
	err := json.Unmarshal(msg.Value, &delayMsg)
	if err != nil {
		return fmt.Errorf("序列化延迟消息失败  %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = receiver.svc.Insert(ctx, delayMsg)
	if err != nil {
		return fmt.Errorf("转储消息失败 %w", err)
	}
	_, err = receiver.consumer.CommitMessage(msg)
	if err != nil {
		return fmt.Errorf("转储消息失败 %w", err)
	}
	return nil
}

// 延迟消息发送者
type DelayMsgSender struct {
	topicConn *syncx.Map[string, *kafka.Producer]
	svc       service.DelayMsgSvc
}

func (sender *DelayMsgSender) SendMsg() {
	// 每30s轮询一次，根据自己的业务进行修改
	tiker := time.NewTicker(30 * time.Second)
	for {
		<-tiker.C
		fmt.Println("轮询中。。。。。")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		sender.sendMsgs(ctx)
		cancel()
	}
}

func (sender *DelayMsgSender) sendMsgs(ctx context.Context) {
	// 获取需要发送的消息
	msgs, err := sender.svc.FindDelayMsg(ctx)
	if err != nil {
		slog.Error("获取延迟消息失败", slog.Any("err", err))
	}
	// 准备发送
	var wg sync.WaitGroup
	for _, msg := range msgs {
		wg.Add(1)
		msg := msg
		go func() {
			defer wg.Done()
			newctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			eerr := sender.sendMsg(newctx, msg)
			cancel()
			if eerr != nil {
				slog.Error("转发延迟消息失败", slog.Any("err", err))
			} else {
				slog.Info("成功转发延迟消息", slog.Any("msg", msg.Value))
			}
		}()
	}
	wg.Wait()
}

func (sender *DelayMsgSender) sendMsg(ctx context.Context, msg domain.DelayMsg) error {
	conn, ok := sender.topicConn.Load(msg.Topic)
	if !ok {
		return fmt.Errorf("未知topic %s", msg.Topic)
	}
	topic := msg.Topic
	err := conn.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic},
		Value:          []byte(msg.Value),
	}, nil)
	if err != nil {
		return fmt.Errorf("消息 %v 转发失败 %w", msg, err)
	}
	err = sender.svc.Complete(ctx, []int64{msg.ID})
	if err != nil {
		return fmt.Errorf("更新消息%v 失败 %w", msg, err)
	}
	return nil
}

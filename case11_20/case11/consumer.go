package case11

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ecodeclub/ekit/syncx"
	kafkago "github.com/segmentio/kafka-go"
	"log"
	"log/slog"
	"time"
)

type DelayConsumer struct {
	// 这部分信息通过配置文件获取
	partitionMap *syncx.Map[int, time.Duration]
	// 这部分提前建立需要转发业务topic及其对应连接，可以通过配置文件获取
	topicMap *syncx.Map[string, *kafkago.Writer]
	reader   *kafkago.Reader
}

type DelayMsg struct {
	Topic string `json:"topic"`
	// 真正需要转发的数据
	Data string `json:"data"`
}

func (d *DelayConsumer) Consume() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		msg, err := d.reader.ReadMessage(ctx)
		cancel()
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			continue
		}
		if err != nil {
			slog.Error("获取延迟消息失败", slog.Any("err", err))
		}
		log.Printf("获取到消息 %v", msg)
		err = d.consume(msg)
		if err != nil {
			slog.Error("转发延迟消息失败", slog.Any("err", err))
		}
	}
}

func (d *DelayConsumer) consume(msg kafkago.Message) error {
	var delayMsg DelayMsg
	err := json.Unmarshal(msg.Value, &delayMsg)
	if err != nil {
		return fmt.Errorf("延迟消息格式错误，序列化失败 %v", err)
	}
	// 根据分区id找寻睡眠时间
	sleepTime, _ := d.partitionMap.Load(msg.Partition)
	timer := time.NewTimer(sleepTime)
	<-timer.C
	// 进行转发
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = d.sendMsg(ctx, delayMsg)
	if err != nil {
		return err
	}
	// 转发成功后才提交，保证消息的一致性。
	err = d.reader.CommitMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("提交消息失败 offset %d Topic %s, 原因 %w", msg.Offset, msg.Topic, err)
	}
	return nil
}

func (d *DelayConsumer) sendMsg(ctx context.Context, delayMsg DelayMsg) error {
	topicWriter, ok := d.topicMap.Load(delayMsg.Topic)
	if !ok {
		return fmt.Errorf("未知topic，%s", delayMsg.Topic)
	}
	err := topicWriter.WriteMessages(ctx, kafkago.Message{
		Value: []byte(delayMsg.Data),
	})
	if err != nil {
		return fmt.Errorf("延迟消息转发失败 %v", err)
	}
	return nil
}

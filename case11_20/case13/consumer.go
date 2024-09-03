package case13

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ecodeclub/ekit/syncx"
	"log"
	"log/slog"
	"time"
)

const (
	delayConsumerGroupName = "delayConsumerGroup"
	defaultPollInterval    = 1 * time.Second
)

type DelayConsumer struct {
	consumer *kafka.Consumer
	// 记录分区和超时时间的关系 可以从配置文件中获取
	partitionMap *syncx.Map[int, time.Duration]
	// 记录topic和其kafka连接
	topicConn *syncx.Map[string, *kafka.Producer]
}

type DelayMsg struct {
	// 转发的数据
	Data string `json:"data"`
	// 转发的主题
	Topic string `json:"topic"`
}

func NewDelayConsumer(addr string, topicMap *syncx.Map[string, *kafka.Producer], partitionMap *syncx.Map[int, time.Duration]) (*DelayConsumer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":  addr,
		"group.id":           delayConsumerGroupName,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	}
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	err = consumer.SubscribeTopics([]string{delayTopic}, nil)
	if err != nil {
		return nil, err
	}
	return &DelayConsumer{
		consumer:     consumer,
		topicConn:    topicMap,
		partitionMap: partitionMap,
	}, nil
}

func (d *DelayConsumer) Consume() {
	for {
		msg, err := d.consumer.ReadMessage(-1)
		if err != nil {
			// 失败记录一下报错然后重试
			slog.Error("获取延迟消息失败", slog.Any("err", err))
			continue
		}
		vv, _ := json.Marshal(msg)
		slog.Info("成功获取延迟消息", slog.Any("msg", string(vv)))
		// 转发
		err = d.consume(msg)
		if err != nil {
			slog.Error("转发延迟消息失败", slog.Any("err", err))
		}
	}
}

func (d *DelayConsumer) consume(msg *kafka.Message) error {
	// 获取发送时间
	sendTime := msg.Timestamp
	now := time.Now()
	// 获取当前分区需要睡多久
	interval, ok := d.partitionMap.Load(int(msg.TopicPartition.Partition))
	if !ok {
		return fmt.Errorf("未知延迟分区")
	}
	log.Printf("msg %v 睡眠分区为 %f 分钟", string(msg.Value), interval.Minutes())
	// 计算要睡多久
	subTime := sendTime.Add(interval).Sub(now)
	log.Printf("msg %v 要睡 %f分钟", string(msg.Value), subTime.Minutes())
	if subTime > 0 {
		// 暂停分区消费
		err := d.consumer.Pause([]kafka.TopicPartition{msg.TopicPartition})
		if err != nil {
			return fmt.Errorf("暂停分区失败 %v", err)
		}
		// 睡眠
		d.sleep(subTime)
		// 恢复分区消费
		err = d.consumer.Resume([]kafka.TopicPartition{msg.TopicPartition})
		if err != nil {
			return fmt.Errorf("恢复分区失败 %v", err)
		}
	}
	// 转发
	err := d.sendMsg(msg)
	if err != nil {
		return err
	}
	_, err = d.consumer.CommitMessage(msg)
	if err != nil {
		return fmt.Errorf("提交消息失败 offset %d Topic %s, 原因 %w", msg.TopicPartition.Offset, delayTopic, err)
	}
	return nil
}

func (d *DelayConsumer) sleep(subTime time.Duration) {
	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()
	timer := time.NewTimer(subTime)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return
		case <-ticker.C:
			d.consumer.Poll(100)
		}
	}
}

func (d *DelayConsumer) sendMsg(msg *kafka.Message) error {
	var delayMsg DelayMsg
	err := json.Unmarshal(msg.Value, &delayMsg)
	if err != nil {
		return fmt.Errorf("延迟消息序列化失败 %v", err)
	}
	topic := delayMsg.Topic
	// 获取转发producer
	producer, ok := d.topicConn.Load(topic)
	if !ok {
		return fmt.Errorf("未知topic %s", topic)
	}
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic},
		Value:          []byte(delayMsg.Data),
	}, nil)
	if err != nil {
		return fmt.Errorf("转发失败 %v", err)
	}
	return nil
}

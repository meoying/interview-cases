package consumer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"time"
)

const bizTopic = "biz_topic"

type BizConsumer struct {
	consumer *kafka.Consumer
}

// NewBizConsumer 模拟业务消费者
func NewBizConsumer(addr string) (*BizConsumer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": addr,
		"auto.offset.reset": "earliest",
		"group.id":          "biz_group",
	}
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}
	err = consumer.SubscribeTopics([]string{bizTopic}, nil)
	if err != nil {
		return nil, err
	}
	return &BizConsumer{
		consumer: consumer,
	}, nil
}

func (b *BizConsumer) Consume(timeout time.Duration) (string, error) {
	msg, err := b.consumer.ReadMessage(timeout)
	if err != nil {
		return "", err
	}
	return string(msg.Value), nil
}

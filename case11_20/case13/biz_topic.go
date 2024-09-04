package case13

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const bizTopic =  "biz_topic"

type BizConsumer struct {
	consumer *kafka.Consumer
}

func NewBizConsumer(addr string) (*BizConsumer,error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": addr,
		"auto.offset.reset": "earliest",
		"group.id": "biz_group",
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
	},nil
}

func (b *BizConsumer) Consume() (string, error) {
	msg,err := b.consumer.ReadMessage(-1)
	if err != nil {
		return "", err
	}
	return string(msg.Value), nil
}

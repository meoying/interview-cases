package producer

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Producer 业务方使用的 producer
type Producer struct {
	producer *kafka.Producer
}

func NewProducer(producer *kafka.Producer) *Producer {
	return &Producer{producer}
}

const delayTopic = "delay_topic"

func (p *Producer) Produce(ctx context.Context,
	msg []byte,
	// 时间戳，毫秒数
	deadline int64,
	bizTopic string) error {
	topic := delayTopic
	delayMsg := DelayMsg{
		Value:    msg,
		Topic:    bizTopic,
		Deadline: deadline,
	}
	msgByte, err := json.Marshal(delayMsg)
	if err != nil {
		return err
	}
	return p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic: &topic,
		},
		Value: msgByte,
	}, nil)
}

package case14

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"interview-cases/case11_20/case14/domain"
)


type Producer struct {
	producer     *kafka.Producer
}
func (p *Producer) Produce(ctx context.Context, msg domain.DelayMsg) error {
	topic := delayTopic
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
		},
		Value: msgByte,
	}, nil)
}

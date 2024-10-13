package case14

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ecodeclub/ekit/syncx"
	"time"
)

const delayTopic = "delay_topic"

type Producer struct {
	// 记录分区和超时时间的关系
	partitionMap *syncx.Map[time.Duration, int]
	producer     *kafka.Producer
}

func (p *Producer) Produce(ctx context.Context, msg DelayMsg, delayTime time.Duration) error {
	topic := delayTopic
	partition, ok := p.partitionMap.Load(delayTime)
	if !ok {
		return errors.New("不支持的超时时间")
	}
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: int32(partition),
		},
		Value: msgByte,
	}, nil)
}

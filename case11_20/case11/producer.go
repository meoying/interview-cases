package case11

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ecodeclub/ekit/syncx"
	"github.com/segmentio/kafka-go"
	"time"
)

const delayTopic = "delay_topic"

type Producer struct {
	writer       *kafka.Writer
	partitionMap syncx.Map[time.Duration, int]
}

func (p *Producer) Produce(ctx context.Context, message DelayMsg, sleepTime time.Duration) error {
	partitition, ok := p.partitionMap.Load(sleepTime)
	if !ok {
		return fmt.Errorf("不支持的超时时间")
	}
	delayMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("延迟消息序列化失败")
	}

	msg := kafka.Message{
		Value:     delayMsg,
		Partition: partitition,
		WriterData: metaMessage{
			SpecifiedPartitionKey: partitition,
		},
	}
	return p.writer.WriteMessages(ctx, msg)
}

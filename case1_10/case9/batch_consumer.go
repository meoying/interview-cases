package case9

import (
	"context"
	kafkago "github.com/segmentio/kafka-go"
	"log"
	"time"
)

type BatchConsumer struct {
	reader *kafkago.Reader
}

func NewBatchConsumer(groupId string) *BatchConsumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   topic,
		GroupID: groupId,
	})
	return &BatchConsumer{reader: reader}
}

func (c *BatchConsumer) Consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		msgs := make([]kafkago.Message, 0, BatchMaxSize)
		// 获取一批数据
		for i := 0; i < BatchMaxSize; i++ {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Println(err)
				return
			}
			msgs = append(msgs, msg)
		}
		// 批量消费
		BatchTest()
		err := c.reader.CommitMessages(ctx, msgs...)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func BatchConsume() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	consumer := NewBatchConsumer("batch_consume")
	consumer.Consume(ctx)
}

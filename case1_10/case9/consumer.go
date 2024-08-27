package case9

import (
	"context"
	kafkago "github.com/segmentio/kafka-go"
	"log"
	"time"
)

type Consumer struct {
	reader *kafkago.Reader
}

func NewConsumer(groupId string) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   topic,
		GroupID: groupId,
	})
	return &Consumer{reader: reader}
}

func (c *Consumer) Consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Println(err)
			return
		}
		// 单个消费
		Test()
		err = c.reader.CommitMessages(ctx, msg)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func Consume() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	consumer := NewConsumer("consume")
	consumer.Consume(ctx)
}

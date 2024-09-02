package case11

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"
)

const bizTopic =  "biz_topic"

type BizConsumer struct {
	topic  string
	reader *kafkago.Reader
}

func NewBizConsumer(topic string, breakers []string) *BizConsumer {
	return &BizConsumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Topic:   topic,
			Brokers: breakers,
			GroupID: "biz_consumer",
		}),
		topic: topic,
	}
}

func (b *BizConsumer) Consume() (string, error) {
	msg, err := b.reader.ReadMessage(context.Background())
	if err != nil {
		return "", err
	}
	delayMsg := msg.Value
	return string(delayMsg), nil
}

package case9

import (
	"context"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"log"
	"time"
)

// 生产者
const topic = "test"

type TestProducer struct {
	writer *kafkago.Writer
}

func (t *TestProducer) Produce(ctx context.Context) {
	val := []byte(fmt.Sprintf("%d", time.Now().UnixMilli()))
	msg := kafkago.Message{
		Key:   []byte("test"),
		Value: val,
	}
	err := t.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("发送消息失败 %v", err)
		return
	}
}

func NewTestProducer() *TestProducer {
	// 配置Kafka写入器
	writer := &kafkago.Writer{
		Addr:     kafkago.TCP("localhost:9092"),
		Topic:    topic,
		Balancer: &kafkago.Hash{},
	}
	return &TestProducer{
		writer: writer,
	}
}

func Produce() {
	producer := NewTestProducer()
	for {
		time.Sleep(1*time.Millisecond)
		producer.Produce(context.Background())
	}
}
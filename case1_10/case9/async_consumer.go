package case9

import (
	"context"
	kafkago "github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"log"
	"time"
)

type AsyncConsumer struct {
	reader *kafkago.Reader
}

func NewAsyncConsumer(groupId string) *AsyncConsumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   topic,
		GroupID: groupId,
	})
	return &AsyncConsumer{reader: reader}
}

func (a *AsyncConsumer) Consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		msgs := make([]kafkago.Message, 0, BatchMaxSize)
		// 获取一批数据
		for i := 0; i < BatchMaxSize; i++ {
			msg, err := a.reader.ReadMessage(ctx)
			if err != nil {
				log.Println(err)
				return
			}
			msgs = append(msgs, msg)
		}
		// 批量消费
		var eg errgroup.Group
		for i:= 0; i < BatchMaxSize; i++ {
			eg.Go(func() error {
				Test()
				return nil
			})
		}
		err := eg.Wait()
		if err != nil {
			return
		}
		err = a.reader.CommitMessages(ctx, msgs...)
		if err != nil {
			log.Println(err)
			return
		}
	}
}


func AsyncConsume(){
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	consumer := NewAsyncConsumer("async_consume")
	consumer.Consume(ctx)
}
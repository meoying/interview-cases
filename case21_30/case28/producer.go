package case28

import (
	"context"
	"encoding/json"
	"github.com/ecodeclub/ekit/slice"
	kafkago "github.com/segmentio/kafka-go"
)

type AggMsg struct {
	Ids []int64 `json:"ids"`
}

type Msg struct {
	ID int64 `json:"id"`
}

const (
	AggMsgTopic = "agg_msg_topic"
	MsgTopic = "msg_topic"
)

type AggProducer struct {
	writer   *kafkago.Writer
}

func NewAggProducer(writer *kafkago.Writer) *AggProducer {
	return &AggProducer{writer}
}

// ProduceAggMsg 发送聚合消息
func (p *AggProducer) ProduceAggMsg(ctx context.Context, msgs []AggMsg) error {
	kafkaMsgs := slice.Map(msgs, func(idx int, src AggMsg) kafkago.Message {
		msgByte,_ := json.Marshal(src)
		return kafkago.Message{
			Topic: AggMsgTopic,
			Value: msgByte,
		}
	})
	return p.writer.WriteMessages(ctx, kafkaMsgs...)
}

// ProduceMsg 发送单条消息
func (p *AggProducer) ProduceMsg(ctx context.Context, msgs []Msg) error {
	kafkaMsgs := slice.Map(msgs, func(idx int, src Msg) kafkago.Message {
		msgByte,_ := json.Marshal(src)
		return kafkago.Message{
			Topic: MsgTopic,
			Value: msgByte,
		}
	})
	return p.writer.WriteMessages(ctx, kafkaMsgs...)
}

package case11

import (
	"github.com/ecodeclub/ekit/syncx"
	kafkago "github.com/segmentio/kafka-go"
	"time"
)

type DelayConsumerGroup struct {
	// 这部分信息通过配置文件获取
	partitionMap *syncx.Map[int, time.Duration]
	// 这部分提前建立需要转发业务topic及其对应连接，可以通过配置文件获取
	topicMap *syncx.Map[string, *kafkago.Writer]
}

func (d *DelayConsumerGroup) Name() string {
	return "delay_group"
}

func (d *DelayConsumerGroup) newConsumer(brokers []string) *DelayConsumer {
	consumer := &DelayConsumer{
		partitionMap: d.partitionMap,
		topicMap:     d.topicMap,
	}
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: brokers,
		Topic:   delayTopic,
		GroupID: d.Name(),
	})
	consumer.reader = reader
	return consumer
}

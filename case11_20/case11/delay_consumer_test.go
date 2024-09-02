package case11

import (
	"context"
	"github.com/ecodeclub/ekit/syncx"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	brokers  []string
	producer *Producer
}

func (s *TestSuite) SetupSuite() {
	// 初始化prod
	brokers := []string{"localhost:9092"}
	s.brokers = brokers
	s.initTopic()
	s.initProducerAndConsumer()
}
func (s *TestSuite) initTopic() {
	conn, err := kafkago.DialContext(context.Background(), "tcp", s.brokers[0])
	if err != nil {
		log.Fatalf("连接 Kafka 失败: %v", err)
	}
	defer conn.Close()
	// 初始化delay_topic
	delayTopicConfig := kafkago.TopicConfig{
		Topic:             delayTopic,
		NumPartitions:     3,
		ReplicationFactor: 1,
	}
	// 创建 topic
	bizTopicConfig := kafkago.TopicConfig{
		Topic:             bizTopic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}
	s.newTopic(conn, []kafkago.TopicConfig{delayTopicConfig, bizTopicConfig})
}

func (s *TestSuite) newTopic(conn *kafkago.Conn, configs []kafkago.TopicConfig) {
	for _, config := range configs {
		err := conn.CreateTopics(config)
		if err != nil {
			log.Fatalf("创建 topic 失败: %v", err)
		}
		log.Printf("Topic %s 创建成功", config.Topic)
	}
}

func (s *TestSuite) initProducerAndConsumer() {
	balancer, _ := NewSpecifiedPartitionBalancer(&kafkago.Hash{})
	writer := &kafkago.Writer{
		Addr:     kafkago.TCP(s.brokers...),
		Topic:    delayTopic,
		Balancer: balancer,
	}
	producer := &Producer{
		writer: writer,
	}
	s.producer = producer
	// 这部分初始化 可以通过配置文件进行配置
	proMap := syncx.Map[time.Duration, int]{}
	proMap.Store(5*time.Minute, 0)
	proMap.Store(10*time.Minute, 1)
	proMap.Store(20*time.Minute, 2)
	producer.partitionMap = proMap
	consumeMap := syncx.Map[int, time.Duration]{}
	consumeMap.Store(0, 5*time.Minute)
	consumeMap.Store(1, 10*time.Minute)
	consumeMap.Store(2, 20*time.Minute)
	bizWriter := &kafkago.Writer{
		Addr:  kafkago.TCP(s.brokers...),
		Topic: bizTopic,
	}
	topicMap := syncx.Map[string, *kafkago.Writer]{}
	topicMap.Store(bizTopic, bizWriter)
	consumerGroup := &DelayConsumerGroup{
		partitionMap: &consumeMap,
		topicMap:     &topicMap,
	}
	// 启动消费组内所有消费者
	go consumerGroup.newConsumer(s.brokers).Consume()
	go consumerGroup.newConsumer(s.brokers).Consume()
	go consumerGroup.newConsumer(s.brokers).Consume()
}

func (s *TestSuite) TestDelayMsg() {
	bizConsumer := NewBizConsumer(bizTopic, s.brokers)
	err := s.producer.Produce(context.Background(), DelayMsg{
		Topic: bizTopic,
		Data:  "test",
	}, 10*time.Minute)
	require.NoError(s.T(), err)
	log.Println("发送消息成功")
	starTime := time.Now().Unix()
	msg, err := bizConsumer.Consume()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test", msg)
	endTime := time.Now().Unix()
	subTime := endTime - starTime
	// 测试结果，延迟的时间应该  10m左右，
	require.True(s.T(), time.Duration(subTime)*time.Second > 10*time.Minute && time.Duration(subTime)*time.Second < 11*time.Minute)
}

func TestDelayMsg(t *testing.T) {
	suite.Run(t, &TestSuite{
		// 记得换你的 Kafka 地址
		brokers: []string{"localhost:9092"},
	})
}

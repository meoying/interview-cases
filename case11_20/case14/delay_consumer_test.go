package case14

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ecodeclub/ekit/syncx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	addr        string
	producer    *Producer
	bizConsumer *BizConsumer
}

func (s *TestSuite) SetupSuite() {
	s.initTopic()
	s.initProducerAndConsumer()
}

func (s *TestSuite) initProducerAndConsumer() {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": s.addr,
	})
	require.NoError(s.T(), err)
	proMap := syncx.Map[time.Duration, int]{}
	proMap.Store(3*time.Minute, 0)
	proMap.Store(5*time.Minute, 1)
	proMap.Store(10*time.Minute, 2)
	producer := &Producer{
		producer:     kafkaProducer,
		partitionMap: &proMap,
	}
	consumeMap := syncx.Map[int, time.Duration]{}
	consumeMap.Store(0, 3*time.Minute)
	consumeMap.Store(1, 5*time.Minute)
	consumeMap.Store(2, 10*time.Minute)

	topicMap := syncx.Map[string, *kafka.Producer]{}
	topicMap.Store(bizTopic, kafkaProducer)
	consumer0, err := NewDelayConsumer(s.addr, &topicMap, &consumeMap)
	require.NoError(s.T(), err)
	consumer1, err := NewDelayConsumer(s.addr, &topicMap, &consumeMap)
	require.NoError(s.T(), err)
	consumer2, err := NewDelayConsumer(s.addr, &topicMap, &consumeMap)
	require.NoError(s.T(), err)
	// 启动三个消费者
	go consumer0.Consume()
	go consumer1.Consume()
	go consumer2.Consume()
	s.producer = producer
	s.bizConsumer, err = NewBizConsumer(s.addr)
	require.NoError(s.T(), err)
}

func (s *TestSuite) TestDelayConsume() {
	// 发送消息
	startTime1 := s.sendMsg("delayMsg1", 10*time.Minute)
	startTime2 := s.sendMsg("delayMsg2", 3*time.Minute)
	time.Sleep(1 * time.Minute)
	startTime3 := s.sendMsg("delayMsg3", 3*time.Minute)
	startTime4 := s.sendMsg("delayMsg4", 5*time.Minute)
	wantMsgs := []WantDelayMsg{
		{
			StartTime:    startTime2,
			IntervalTime: 3 * time.Minute,
			Data:         "delayMsg2",
		},
		{
			StartTime:    startTime3,
			IntervalTime: 3 * time.Minute,
			Data:         "delayMsg3",
		},
		{
			StartTime:    startTime4,
			IntervalTime: 5 * time.Minute,
			Data:         "delayMsg4",
		},
		{
			StartTime:    startTime1,
			IntervalTime: 10 * time.Minute,
			Data:         "delayMsg1",
		},
	}
	for _, want := range wantMsgs {
		startTime := want.StartTime
		msg, err := s.bizConsumer.Consume()
		subTime := time.Now().Sub(startTime)
		log.Printf("开始校验 %v 睡了 %f分钟", want, subTime.Minutes())
		require.NoError(s.T(), err)
		assert.Equal(s.T(), want.Data, msg)

		// 允许误差10s
		require.True(s.T(), subTime >= want.IntervalTime-10*time.Second && subTime <= want.IntervalTime+10*time.Second)
	}
}

func (s *TestSuite) sendMsg(data string, intervalTime time.Duration) time.Time {
	err := s.producer.Produce(context.Background(), DelayMsg{
		Data:  data,
		Topic: bizTopic,
	}, intervalTime)
	require.NoError(s.T(), err)
	return time.Now()
}

func TestDelayMsg(t *testing.T) {

	suite.Run(t, &TestSuite{
		// 记得换你的 Kafka 地址
		addr: "127.0.0.1:9092",
	})
}

type WantDelayMsg struct {
	StartTime    time.Time
	IntervalTime time.Duration
	Data         string
}

func (s *TestSuite) initTopic() {
	// 创建 AdminClient
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": s.addr,
	})
	require.NoError(s.T(), err)
	defer adminClient.Close()
	// 设置要创建的主题的配置信息
	topic := delayTopic
	numPartitions := 3
	replicationFactor := 1
	// 创建主题
	results, err := adminClient.CreateTopics(
		context.Background(),
		[]kafka.TopicSpecification{
			{
				Topic:             topic,
				NumPartitions:     numPartitions,
				ReplicationFactor: replicationFactor,
			},
			{
				Topic:             bizTopic,
				NumPartitions:     1,
				ReplicationFactor: replicationFactor,
			},
		},
	)
	require.NoError(s.T(), err)
	// 处理创建主题的结果
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("创建topic失败 %s: %v\n", result.Topic, result.Error)

		} else {
			fmt.Printf("Topic %s 创建成功\n", result.Topic)
		}
	}
}

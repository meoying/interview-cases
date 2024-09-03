package case13

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
	//time.Sleep(1 * time.Minute)
	//startTime3 := s.sendMsg("delayMsg3", 3*time.Minute)
	startTime4 := s.sendMsg("delayMsg4", 5*time.Minute)
	wantMsgs := []WantDelayMsg{
		{
			StartTime:    startTime2,
			IntervalTime: 3 * time.Minute,
			Data:         "delayMsg2",
		},
		//{
		//	StartTime:    startTime3,
		//	IntervalTime: 3 * time.Minute,
		//	Data:         "delayMsg3",
		//},
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
		require.True(s.T(), subTime >= want.IntervalTime && subTime <= want.IntervalTime+10*time.Second)
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
		addr: "localhost:9092",
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
		"bootstrap.servers": "localhost:9092",
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

//
//func TestConsumerPauseResume(t *testing.T) {
//	// 创建 Kafka 生产者
//	producer, err := kafka.NewProducer(&kafka.ConfigMap{
//		"bootstrap.servers": "localhost:9092",
//	})
//	if err != nil {
//		t.Fatalf("Failed to create producer: %s", err)
//	}
//	defer producer.Close()
//
//	// 创建 Kafka 消费者
//	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
//		"bootstrap.servers":    "localhost:9092",
//		"group.id":             "test-group",
//		"auto.offset.reset":    "earliest",
//		"enable.auto.commit":   "false", // 关闭自动提交
//		"max.poll.interval.ms": "60000", // 设置为 10 分钟
//		//"session.timeout.ms":   "5000",
//	})
//	if err != nil {
//		t.Fatalf("Failed to create consumer: %s", err)
//	}
//	defer consumer.Close()
//
//	// 订阅主题
//	topic := "test-topic"
//	err = consumer.Subscribe(topic, nil)
//	if err != nil {
//		t.Fatalf("Failed to subscribe to topic: %s", err)
//	}
//
//	// 生产消息到 Kafka
//	for i := 0; i < 100; i++ {
//		msg := &kafka.Message{
//			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
//			Value:          []byte(fmt.Sprintf("%d", i)),
//		}
//		err := producer.Produce(msg, nil)
//		if err != nil {
//			t.Fatalf("Failed to produce message: %s", err)
//		}
//	}
//
//	// 创建通道接收消息
//	msgChan := make(chan *kafka.Message, 10)
//	go func() {
//		for {
//			msg, err := consumer.ReadMessage(-1)
//			if err != nil {
//				t.Errorf("Failed to read message: %s", err)
//				continue
//			}
//			log.Println(string(msg.Value))
//			time.Sleep(100 * time.Millisecond)
//			msgChan <- msg
//		}
//	}()
//
//	// 等待消费者开始处理消息
//	time.Sleep(5 * time.Second)
//
//	// 暂停分区
//	partitions := []kafka.TopicPartition{
//		{Topic: &topic, Partition: 0},
//	}
//	consumer.Pause(partitions)
//	t.Log("Paused partition 0")
//	go func() {
//		for {
//			evt := consumer.Poll(11)
//			time.Sleep(1 * time.Minute)
//			log.Println(evt)
//		}
//	}()
//	// 等待一定时间以观察暂停的效果
//	time.Sleep(6 * time.Minute)
//
//	// 检查在暂停期间是否有消息被消费
//	select {
//	case msg := <-msgChan:
//		t.Errorf("Unexpected message received during pause: %v", msg)
//	default:
//		t.Log("No messages received during pause, as expected")
//	}
//
//	// 恢复分区
//	consumer.Resume(partitions)
//	t.Log("Resumed partition 0")
//
//	// 等待消费者恢复消费
//	time.Sleep(5 * time.Second)
//
//	// 检查恢复后是否有消息被消费
//	for i := 0; i < 1000000; i++ {
//		select {
//		case <-msgChan:
//		case <-time.After(5 * time.Second):
//			t.Errorf("Timed out waiting for message test-message-%d", i)
//		}
//	}
//}

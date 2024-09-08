package case14

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ecodeclub/ekit/syncx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"interview-cases/case11_20/case13"
	"interview-cases/case11_20/case14/domain"
	"interview-cases/case11_20/case14/repository"
	"interview-cases/case11_20/case14/repository/dao"
	"interview-cases/case11_20/case14/service"
	"interview-cases/test"
	"log"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	addr        string
	producer    *Producer
	bizConsumer *case13.BizConsumer
}
type WantDelayMsg struct {
	StartTime    time.Time
	IntervalTime time.Duration
	Data         string
}

func (s *TestSuite) SetupSuite() {
	s.initTopic()
	s.initConsumerAndProducer()
}

// 初始化service
func (s *TestSuite) initSvc() service.DelayMsgSvc {
	db := test.InitDB()
	delayDao, err := dao.NewDelayMsgDAO(db)
	require.NoError(s.T(), err)
	delayRepo := repository.NewDelayMsgRepo(delayDao)
	return service.NewDelayMsgSvc(delayRepo)
}

func (s *TestSuite) initTopic() {
	test.InitTopic(
		kafka.TopicSpecification{
			Topic:         delayTopic,
			NumPartitions: 1,
		},
		kafka.TopicSpecification{
			Topic:         case13.BizTopic,
			NumPartitions: 1,
		},
	)
}

func (s *TestSuite) initConsumerAndProducer() {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": s.addr,
	})
	require.NoError(s.T(), err)
	s.producer = &Producer{
		producer: kafkaProducer,
	}
	svc := s.initSvc()
	config := &kafka.ConfigMap{
		"bootstrap.servers":  s.addr,
		"group.id":           "delay_msg_group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	}
	consumer, err := kafka.NewConsumer(config)
	require.NoError(s.T(), err)
	err = consumer.SubscribeTopics([]string{delayTopic}, nil)
	require.NoError(s.T(), err)
	receiver := DelayMsgReceiver{
		svc:      svc,
		consumer: consumer,
	}
	// 启动消息接收者
	go receiver.ReceiveMsg()
	topicMap := syncx.Map[string, *kafka.Producer]{}
	topicMap.Store(case13.BizTopic, kafkaProducer)
	sender := DelayMsgSender{
		svc:       svc,
		topicConn: &topicMap,
	}
	// 启动消息发送者
	go sender.SendMsg()
	// 初始化业务消费者
	bizConsumer, err := case13.NewBizConsumer(s.addr)
	require.NoError(s.T(), err)
	s.bizConsumer = bizConsumer
}

func (s *TestSuite) sendMsg(data string, intervalTime time.Duration) time.Time {
	deadline := time.Now().Add(intervalTime)
	err := s.producer.Produce(context.Background(), domain.DelayMsg{
		Value:    data,
		Deadline: deadline,
		Topic:    case13.BizTopic,
	})
	require.NoError(s.T(), err)
	return time.Now()
}

func (s *TestSuite) TestDelayMsg() {
	// 发送消息
	startTime1 := s.sendMsg("delayMsg1", 600*time.Second)
	startTime2 := s.sendMsg("delayMsg2", 150*time.Second)
	time.Sleep(1 * time.Minute)
	startTime3 := s.sendMsg("delayMsg3", 250*time.Second)
	startTime4 := s.sendMsg("delayMsg4", 320*time.Second)
	startTime5 := s.sendMsg("delayMsg5", 420*time.Second)
	wantMsgs := []WantDelayMsg{
		{
			StartTime:    startTime2,
			IntervalTime: 150 * time.Second,
			Data:         "delayMsg2",
		},
		{
			StartTime:    startTime3,
			IntervalTime: 250 * time.Second,
			Data:         "delayMsg3",
		},
		{
			StartTime:    startTime4,
			IntervalTime: 320 * time.Second,
			Data:         "delayMsg4",
		},
		{
			StartTime:    startTime5,
			IntervalTime: 420 * time.Second,
			Data:         "delayMsg5",
		},
		{
			StartTime:    startTime1,
			IntervalTime: 600 * time.Second,
			Data:         "delayMsg1",
		},
	}
	for _, want := range wantMsgs {
		startTime := want.StartTime
		msg, err := s.bizConsumer.Consume()
		subTime := time.Now().Sub(startTime)
		log.Printf("开始校验 %v 睡了 %f秒", want, subTime.Seconds())
		require.NoError(s.T(), err)
		assert.Equal(s.T(), want.Data, msg)

		// 允许误差10s
		require.True(s.T(), subTime >= want.IntervalTime-10*time.Second && subTime <= want.IntervalTime+1*time.Minute)
	}
}

func TestDelayMsg(t *testing.T) {

	suite.Run(t, &TestSuite{
		// 记得换你的 Kafka 地址
		addr: "127.0.0.1:9092",
	})
}

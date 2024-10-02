package case15

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ecodeclub/ekit/syncx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"interview-cases/case11_20/case15/biz/consumer"
	"interview-cases/case11_20/case15/biz/producer"
	"interview-cases/case11_20/case15/delay_platform"
	"interview-cases/case11_20/case15/delay_platform/dao"
	"interview-cases/test"
	"log"
	"testing"
	"time"
)

// biz_topic 业务 topic，在这个测试里面，代表的是订单超时未支付
const bizTopic = "biz_topic"

type TestSuite struct {
	suite.Suite
	addr        string
	producer    *producer.Producer
	bizConsumer *consumer.BizConsumer
	db          *gorm.DB
}
type WantDelayMsg struct {
	StartTime    time.Time
	IntervalTime time.Duration
	Data         string
}

func (s *TestSuite) SetupSuite() {
	s.db = test.InitDB()
	s.initTopic()
	s.initConsumerAndProducer()
}

func (s *TestSuite) initTopic() {
	test.InitTopic(
		kafka.TopicSpecification{
			Topic:         "delay_topic",
			NumPartitions: 1,
		},
		kafka.TopicSpecification{
			Topic:         bizTopic,
			NumPartitions: 1,
		},
	)
}

func (s *TestSuite) initConsumerAndProducer() {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": s.addr,
	})
	require.NoError(s.T(), err)
	s.producer = producer.NewProducer(kafkaProducer)
	require.NoError(s.T(), err)
	msgDAO := dao.NewDelayMsgDAO(s.db)

	config := &kafka.ConfigMap{
		"bootstrap.servers":  s.addr,
		"group.id":           "delay_msg_group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	}
	kaCon, err := kafka.NewConsumer(config)
	require.NoError(s.T(), err)
	err = kaCon.SubscribeTopics([]string{"delay_topic"}, nil)
	require.NoError(s.T(), err)
	receiver := delay_platform.NewDelayMsgReceiver(kaCon, msgDAO)
	// 启动延迟消息接收者，测试环境下，一个就够了
	// 在实践中，delay_topic 有多少个分区就有多少个接收者
	go receiver.ReceiveMsg()

	// 启动所有的延迟消息发送者
	s.initSenders(kafkaProducer, msgDAO)

	// 初始化业务消费者
	bizConsumer, err := consumer.NewBizConsumer(s.addr)
	require.NoError(s.T(), err)
	s.bizConsumer = bizConsumer
}

func (s *TestSuite) initSenders(kafkaProducer *kafka.Producer, msgDAO *dao.DelayMsgDAO) {
	topicMap := &syncx.Map[string, *kafka.Producer]{}
	topicMap.Store(bizTopic, kafkaProducer)
	tables := []string{
		"delay_msg_db_0.delay_msg_tab_0",
		"delay_msg_db_0.delay_msg_tab_1",
		"delay_msg_db_1.delay_msg_tab_0",
		"delay_msg_db_1.delay_msg_tab_1"}
	// 每张表启动一个 sender
	// 在实践中，这个地方应该做成抢占式的，
	// 也就是一张表一个分布式锁，谁拿到就谁来发送
	for _, table := range tables {
		sender := delay_platform.NewDelayMsgSender(topicMap, msgDAO, table)
		// 启动延迟消息发送者
		go sender.SendMsg()
	}

}

// sendMsg 模拟业务方发送延迟消息
func (s *TestSuite) sendMsg(data string, intervalTime time.Duration) time.Time {
	deadline := time.Now().Add(intervalTime)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := s.producer.Produce(ctx, []byte(data), deadline.UnixMilli(), bizTopic)
	require.NoError(s.T(), err)
	return time.Now()
}

func (s *TestSuite) TestDelayMsg() {
	// 发送消息
	// 比如说是发送订单超时未支付
	// 用户设置的闹钟或者提醒事项
	// 这个测试运行的时间很长（10分钟），你要耐心等
	// 你也可以把全部的时间都调低，但是要一一对应
	startTime1 := s.sendMsg("delayMsg1", 600*time.Second)
	startTime2 := s.sendMsg("delayMsg2", 150*time.Second)
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
		msg, err := s.bizConsumer.Consume(-1)
		subTime := time.Now().Sub(startTime)
		log.Printf("开始校验 %v 睡了 %f秒", want, subTime.Seconds())
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), want.Data, msg)
		// 允许误差10s
		assert.True(s.T(), subTime >= want.IntervalTime-10*time.Second && subTime <= want.IntervalTime+1*time.Minute)
	}
}

// 如果你要运行这个测试，每次都要把 docker compose 全删了（不是停了），
// 不然你的数据库，Kafka 上有各种残余的数据，你运行会各种失败。
func TestDelayMsg(t *testing.T) {
	suite.Run(t, &TestSuite{
		// 记得换你的 Kafka 地址
		addr: "127.0.0.1:9092",
	})
}

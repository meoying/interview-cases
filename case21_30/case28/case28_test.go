package case28

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
	"interview-cases/test"
	"log"
	"log/slog"
	"math/rand/v2"
	"testing"
	"time"
)

// 场景描述：订单业务，用户付完钱了，订单改成订单已处理状态。然后发送一条消息通知用户订单支付成功准备发货。
// 测试流程：
// 1. 两个生产者，模拟扣款后发送消息到kafka让通知系统去给用户发送消息。一个生产者生产聚合消息。一个生产者生产普通消息
// 2. 两个生产者都分别处理了一万个订单数据
// 3. 两个消费者，去消费kafka中的数据。比较两个消费者消费完这一万个订单数据的时间

type TestSuite struct {
	suite.Suite
	brokers     []string
	orderDao    OrderDAO
	producer    *AggProducer
	aggConsumer *AggConsumer
	consumer    *Consumer
}

func (s *TestSuite) SetupSuite() {
	// 初始化服务
	db := test.InitDB()
	err := InitTables(db)
	require.NoError(s.T(), err)
	test.InitTopic(
		kafka.TopicSpecification{
			Topic:         AggMsgTopic,
			NumPartitions: 1,
		},
		kafka.TopicSpecification{
			Topic:         MsgTopic,
			NumPartitions: 1,
		},
	)
	notifyDao := NewNotificationDAO(db)
	orderDao := NewOrderDAO(db)
	s.producer = NewAggProducer(&kafkago.Writer{
		Addr:     kafkago.TCP(s.brokers...),
		Balancer: &kafkago.Hash{},
	})
	aggReader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: s.brokers,
		Topic:   AggMsgTopic,
		GroupID: "test_group",
	})
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: s.brokers,
		Topic:   MsgTopic,
		GroupID: "test_group_1",
	})
	s.aggConsumer = NewAggConsumer(aggReader, orderDao, notifyDao)
	s.consumer = NewConsumer(reader, orderDao, notifyDao)
	s.orderDao = orderDao
	// 初始化订单数据
	batchSize := 200
	totalOrders := 20000
	// 逐批生成订单并插入数据库
	for i := 1; i <= 20000; i += batchSize {
		var orders []Order
		// 生成一个批次的订单
		for j := 0; j < batchSize && (i+j) <= totalOrders; j++ {
			order := Order{
				ID:          int64(i + j),
				UserID:      int64(rand.IntN(20)),
				ProductID:   int64(i + j),
				ProductName: fmt.Sprintf("商品%d", i+j),
				Quantity:    1,
			}
			orders = append(orders, order)
		}
		// 插入当前批次
		err := orderDao.BatchCreate(context.Background(), orders)
		require.NoError(s.T(), err)
	}
}

func (s *TestSuite) TestAggMsg() {
	var eg errgroup.Group
	// 生产聚合消息
	eg.Go(func() error {
		log.Println("开始生产聚合消息")
		batchSize := 20
		msgs := make([]AggMsg, 0, batchSize)
		for i := 1; i <= 10000; i += batchSize {
			orderIds := make([]int64, 0, batchSize)
			for j := 0; j < batchSize && (i+j) <= 10000; j++ {
				// 这里模拟处理订单
				err := s.orderDao.UpdateStatus(context.Background(), int64(i+j), 1)
				require.NoError(s.T(), err)
				orderIds = append(orderIds, int64(j))
			}
			msgs = append(msgs, AggMsg{
				Ids: orderIds,
			})
		}
		// 发送聚合消息
		err := s.producer.ProduceAggMsg(context.Background(), msgs)
		require.NoError(s.T(), err)
		log.Println("生产聚合消息完毕")
		return nil
	})
	// 生产普通消息
	eg.Go(func() error {
		log.Println("开始生产普通消息")
		batchSize := 1000
		for i := 10001; i <= 20000; i += batchSize {
			// 发普通消息
			msgs := make([]Msg, 0, batchSize)
			for j := 0; j < batchSize && (i+j) <= 20000; j++ {
				// 这里模拟处理订单
				err := s.orderDao.UpdateStatus(context.Background(), int64(i+j), 1)
				require.NoError(s.T(), err)
				msgs = append(msgs, Msg{
					ID: int64(i + j),
				})
			}
			err := s.producer.ProduceMsg(context.Background(), msgs)
			require.NoError(s.T(), err)
		}
		log.Println("普通消息生产完毕")
		return nil
	})
	require.NoError(s.T(), eg.Wait())
	// 消费聚合消息
	eg.Go(func() error {
		startTime := time.Now().UnixMilli()
		s.aggConsumer.AggConsume(context.Background())
		endTime := time.Now().UnixMilli()
		slog.Info(fmt.Sprintf("使用聚合消息处理10000条订单数据需要 %dms", endTime-startTime))
		return nil
	})
	// 消费普通消息
	eg.Go(func() error {
		startTime := time.Now().UnixMilli()
		s.consumer.Consume(context.Background())
		endTime := time.Now().UnixMilli()
		slog.Info(fmt.Sprintf("不使用聚合消息处理10000条订单数据需要 %dms", endTime-startTime))
		return nil
	})
	require.NoError(s.T(), eg.Wait())
}

func TestBatchConsumer(t *testing.T) {
	suite.Run(t, &TestSuite{
		// 记得换你的 Kafka 地址
		brokers: []string{"localhost:9092"},
	})
}

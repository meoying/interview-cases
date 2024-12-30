package case28

import (
	"context"
	"encoding/json"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"log/slog"
	"time"
)

// Consumer 用于订单处理完了给消费者发送消息
type AggConsumer struct {
	notificationDao NotificationDAO
	orderDao        OrderDAO
	reader          *kafkago.Reader
}

func NewAggConsumer(reader *kafkago.Reader, orderDao OrderDAO, notificationDao NotificationDAO) *AggConsumer {
	return &AggConsumer{
		notificationDao: notificationDao,
		orderDao:        orderDao,
		reader:          reader,
	}
}

// AggConsume 处理聚合消息
func (a *AggConsumer) AggConsume(ctx context.Context) {
	var count int
	for count < 10000 {
		if ctx.Err() != nil {
			slog.Error("退出消费循环", slog.Any("err", ctx.Err()))
			return
		}
		msg, err := a.reader.ReadMessage(ctx)
		if err != nil {
			slog.Error("读取消息失败", slog.Any("err", err))
			continue
		}
		// 处理聚合信息
		solveMsgCount, err := a.doAggBiz(msg)
		if err != nil {
			slog.Error("业务处理失败", slog.Any("err", err))
		}
		count+=solveMsgCount
	}
}

func (a *AggConsumer) doAggBiz(kafkaMsg kafkago.Message) (int,error) {
	var msg AggMsg
	err := json.Unmarshal(kafkaMsg.Value, &msg)
	if err != nil {
		return 0,fmt.Errorf("序列化消息失败 %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	orderMap, err := a.orderDao.BatchGetOrder(ctx, msg.Ids)
	if err != nil {
		return 0,fmt.Errorf("批量获取订单失败 %w", err)
	}
	notifications := make([]*Notification, 0)
	// notifyCation 新建通知
	for _, order := range orderMap {
		orderNotify := &Notification{
			UserID:  order.UserID,
			Biz:     "order",
			BizID:   order.ID,
			Message: fmt.Sprintf("你已经购买 %s商品,请耐心等待发货", order.ProductName),
			Type:    "order_notification",
		}
		notifications = append(notifications, orderNotify)
	}
	return len(msg.Ids), a.notificationDao.BatchCreate(ctx, notifications)
}

type Consumer struct {
	notificationDao NotificationDAO
	orderDao        OrderDAO
	reader          *kafkago.Reader
}

func NewConsumer(reader *kafkago.Reader, orderDao OrderDAO, notificationDao NotificationDAO) *Consumer {
	return &Consumer{
		notificationDao: notificationDao,
		orderDao:        orderDao,
		reader:          reader,
	}
}

// Consume 处理10000普通消息
func (c *Consumer) Consume(ctx context.Context) {
	count := 0
	for count < 10000 {
		if ctx.Err() != nil {
			slog.Error("退出消费循环", slog.Any("err", ctx.Err()))
			return
		}
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			slog.Error("读取消息失败", slog.Any("err", err))
			continue
		}
		// 处理聚合信息
		err = c.doBiz(msg)
		if err != nil {
			slog.Error("业务处理失败", slog.Any("err", err))
		}
		count++
	}
}

func (c *Consumer) doBiz(kafkaMsg kafkago.Message) error {
	var msg Msg
	err := json.Unmarshal(kafkaMsg.Value, &msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败 %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	order, err := c.orderDao.GetOrder(ctx, msg.ID)
	if err != nil {
		return fmt.Errorf("批量获取订单失败 %w", err)
	}
	orderNotify := &Notification{
		UserID:  order.UserID,
		Biz:     "order",
		BizID:   order.ID,
		Message: fmt.Sprintf("你已经购买 %s商品,请耐心等待发货", order.ProductName),
		Type:    "order_notification",
	}
	return c.notificationDao.Create(ctx, orderNotify)
}

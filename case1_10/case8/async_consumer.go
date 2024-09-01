package case8

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	kafkago "github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type AsyncConsumer struct {
	reader    *kafkago.Reader
	batchSize int
}

func NewAsyncConsumer(reader *kafkago.Reader, batchSize int) *AsyncConsumer {
	// batchSize 你也可以做成参数
	return &AsyncConsumer{reader: reader, batchSize: batchSize}
}

func (a *AsyncConsumer) Consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			slog.Error("退出消费循环", slog.Any("err", ctx.Err()))
			return
		}
		err := a.batchAsyncConsume(ctx)
		if err != nil {
			slog.Error("消费失败", slog.Any("err", err))
		}
	}
}

// 消费一批
func (a *AsyncConsumer) batchAsyncConsume(ctx context.Context) error {
	var lastMsg kafkago.Message
	// 异步消费
	var eg errgroup.Group
	// 获取一批数据
	// 要注意，如果你的并发不够，你可能很难凑够一批，所以要加上超时控制
	// 举个极端例子，你可能已经异步消费了 3 条数据，但是一两个小时都没等到更多的消息，
	// 这个时候你不能说这三条你就不提交了
	// 我们这里认为一秒钟内要么凑够一批，要么我们就先处理这些
	batchCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	for i := 0; i < a.batchSize; i++ {
		msg, err := a.reader.ReadMessage(batchCtx)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			// 没有凑够一批，但是还是要考虑提交，也就是不要等后面的消息了
			break
		}
		if err != nil {
			return fmt.Errorf("获取消息失败 %w", err)
		}
		lastMsg = msg
		eg.Go(func() error {
			err1 := a.doBiz(msg)
			if err1 != nil {
				return fmt.Errorf("执行业务失败 offset %d, topic %s, 原因 %w", msg.Offset, msg.Topic, err1)
			}
			return nil
		})
	}

	// 说明在 1 秒钟之内，一条数据都没有获取到，没关系，可以尝试获取下一批
	if lastMsg.Offset == 0 {
		return nil
	}

	err := eg.Wait()
	if err != nil {
		return err
	}
	// 在这里可以只提交偏移量最大的，也就是最后一条
	// 因为 Kafka 的特性是你提交了后面的，就认为前面的也被消费了
	err = a.reader.CommitMessages(ctx, lastMsg)
	if err != nil {
		return fmt.Errorf("提交消息失败 offset %d topic %s, 原因 %w", lastMsg.Offset, lastMsg.Topic, err)
	}
	return nil
}

// 执行业务逻辑
func (a *AsyncConsumer) doBiz(msg kafkago.Message) error {
	// 在实践中，这部分可能是发起 rpc 调用
	// 也可能是发起 http 调用
	// 也可能就是自己执行业务逻辑，例如插入数据库
	// 这里我们假设是发起 http 请求
	resp, err := http.Post(
		"http://localhost:8080/handle", "application/json", bytes.NewBuffer(msg.Value))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	slog.Debug("处理完毕", slog.String("resp", string(respBody)))
	return nil
}

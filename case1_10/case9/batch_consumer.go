package case9

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	kafkago "github.com/segmentio/kafka-go"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type BatchConsumer struct {
	reader    *kafkago.Reader
	batchSize int
}

func NewBatchConsumer(reader *kafkago.Reader, batchSize int) *BatchConsumer {
	return &BatchConsumer{reader: reader, batchSize: 5}
}

func (c *BatchConsumer) Consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		err := c.batchConsume(ctx)
		if err != nil {
			slog.Error("消费失败", slog.Any("err", err))
			return
		}
	}
}

func (c *BatchConsumer) batchConsume(ctx context.Context) error {
	batchCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	msgs := make([]kafkago.Message, 0, c.batchSize)
	// 获取一批数据

	for i := 0; i < c.batchSize; i++ {
		msg, err := c.reader.ReadMessage(batchCtx)
		if err != nil {
			// 取出来多少就处理多少
			break
		}
		msgs = append(msgs, msg)
	}
	if len(msgs) == 0 {
		return nil
	}
	// 批量消费
	err := c.batchBiz(msgs)
	if err != nil {
		return fmt.Errorf("批量消费消息失败 %w", err)
	}
	err = c.reader.CommitMessages(ctx, msgs...)
	if err != nil {
		return fmt.Errorf("提交消息失败 %w", err)
	}
	return nil
}

func (c *BatchConsumer) batchBiz(msgs []kafkago.Message) error {
	// 从实际业务上来说，你这里可能是调用 rpc，也可能是调用 Http，或者直接处理
	// 这里我们用 http 接口来模拟。
	// 直接调用批量接口
	vals := slice.Map(msgs, func(idx int, src kafkago.Message) string {
		return string(src.Value)
	})
	data, _ := json.Marshal(vals)
	resp, err := http.Post(
		"http://localhost:8080/batch", "application/json", bytes.NewBuffer(data))
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

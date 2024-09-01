package case8

import (
	"bytes"
	"context"
	kafkago "github.com/segmentio/kafka-go"
	"io"
	"log/slog"
	"net/http"
)

type SyncConsumer struct {
	reader *kafkago.Reader
}

func NewSyncConsumer(reader *kafkago.Reader) *SyncConsumer {
	// batchSize 你也可以做成参数
	return &SyncConsumer{reader: reader}
}

func (a *SyncConsumer) Consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			slog.Error("退出消费循环", slog.Any("err", ctx.Err()))
			return
		}
		msg, err := a.reader.ReadMessage(ctx)
		if err != nil {
			slog.Error("读取消息失败", slog.Any("err", err))
			continue
		}
		err = a.doBiz(msg)
		if err != nil {
			slog.Error("业务处理失败", slog.Any("err", err))
		}
	}
}

// 执行业务逻辑
func (a *SyncConsumer) doBiz(msg kafkago.Message) error {
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

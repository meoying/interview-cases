// Copyright 2023 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package case8

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ecodeclub/ekit/randx"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type Case8TestSuite struct {
	suite.Suite
	brokers []string
	topic   string
}

func (s *Case8TestSuite) SetupSuite() {
	// 在这里启动了生产者
	producer := &kafkago.Writer{
		Addr:     kafkago.TCP(s.brokers...),
		Topic:    s.topic,
		Balancer: &kafkago.Hash{},
	}

	producer.AllowAutoTopicCreation = true
	go func() {
		// 启动业务服务器
		StartServer(":8080")
	}()

	// 生产一百条消息做预备，当然在现实中这个是源源不绝的
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
	defer cancel()
	now := time.Now().UnixMilli()
	msgs := make([]kafkago.Message, 0, 100)
	for i := 0; i < 100; i++ {
		id := now + int64(i)
		name, _ := randx.RandCode(8, randx.TypeMixed)
		user := UserCase8{
			ID:        id,
			Name:      name,
			Email:     fmt.Sprintf("%d@qq.com", id),
			UpdatedAt: now,
			CreatedAt: now,
		}
		val, _ := json.Marshal(user)
		msgs = append(msgs, kafkago.Message{
			Key:   []byte(fmt.Sprintf("%d", id)),
			Value: val,
		})
	}
	err := producer.WriteMessages(ctx, msgs...)
	assert.NoError(s.T(), err)
}

func (s *Case8TestSuite) TestAsyncConsume() {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: s.brokers,
		Topic:   s.topic,
		GroupID: "test_group",
	})
	// 批次大小也会影响性能
	const batchSize = 10
	s.T().Log("开始消费")
	consumer := NewAsyncConsumer(reader, batchSize)
	// 十秒之后就会退出消费。正常在生产环境是不会用 WithTimeout，而是 WithCancel
	// 而后在应用退出的时候 cancel 掉
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	consumer.Consume(ctx)
	s.T().Log("结束消费")
}

func TestAsyncConsumer(t *testing.T) {
	suite.Run(t, &Case8TestSuite{
		// 记得换你的 Kafka 地址
		brokers: []string{"localhost:9092"},
		topic:   "case8_user",
	})
}

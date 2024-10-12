package test

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"

	"time"
)

func InitTopic(topics ...kafka.TopicSpecification) {
	// 创建 AdminClient
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "127.0.0.1:9092",
	})
	if err != nil {
		panic(fmt.Sprintf("创建kafka连接失败: %v", err))
	}
	defer adminClient.Close()
	// 设置要创建的主题的配置信息
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 创建主题
	results, err := adminClient.CreateTopics(
		ctx,
		topics,
	)
	if err != nil {
		panic(fmt.Sprintf("创建topic失败: %v", err))
	}

	// 处理创建主题的结果
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("创建topic失败 %s: %v\n", result.Topic, result.Error)
		} else {
			fmt.Printf("topic %s 创建成功\n", result.Topic)
		}
	}
}

package case9

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
)

func InitTopic() {
	// Kafka broker 地址
	kafkaAddress := "localhost:9092" // 替换为你的 Kafka broker 地址

	// 创建 Kafka 连接
	conn, err := kafka.DialContext(context.Background(), "tcp", kafkaAddress)
	if err != nil {
		log.Fatalf("连接 Kafka 失败: %v", err)
	}
	defer conn.Close()

	// 配置 topic 相关参数
	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}
	// 创建 topic
	err = conn.CreateTopics(topicConfig)
	if err != nil {
		log.Fatalf("创建 topic 失败: %v", err)
	}
	log.Printf("Topic %s 创建成功", topic)
}
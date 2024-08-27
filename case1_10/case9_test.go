package case1_10

import (
	"interview-cases/case1_10/case9"
	"log"
	"testing"
	"time"
)

func TestBatchConsumer(t *testing.T) {
	// 启动下游服务 记录处理了几个数据
	go case9.InitServer()
	// 初始化topic
	case9.InitTopic()
	// 启动生产者
	go case9.Produce()
	go case9.Produce()
	go case9.Produce()
	time.Sleep(60 * time.Second)
	// 启动顺序消费者,会持续消费60s的数据
	case9.Consume()
	// 查看下游服务处理了多少个数据
	count := case9.GetCount()
	log.Printf("逐个消费，处理了多少数据 %d", count)
	// 清理数据
	case9.Clear()
	// 批量消费，也会持续消费60s的数据
	case9.BatchConsume()
	// 查看批量消费处理了多少个数据
	count = case9.GetCount()
	log.Printf("批量消费，处理了多少数据 %d", count)
	case9.Clear()
	case9.AsyncConsume()
	// 查看异步消费处理了多少个数据
	count = case9.GetCount()
	log.Printf("异步消费，处理了多少数据 %d", count)
}

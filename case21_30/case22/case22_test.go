package case22

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"interview-cases/case21_30/case22/pb"
	"interview-cases/case21_30/case22/service"
	"math"
	"sort"
	"sync"
	"testing"

	"interview-cases/case21_30/case22/interceptor"
	"interview-cases/case21_30/case22/monitor"
	"log"
	"net"
	"time"
)

type TestSuite struct {
	suite.Suite
	client pb.TestServiceClient
}

func (t *TestSuite) SetupSuite() {
	// 真正使用的时候，替换成其他实现。这里使用mock方便测试
	mockMon := monitor.NewMockMonitor()
	// 初始化grpc服务端,注册拦截器
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.NewMemoryLimiter(mockMon, 1*time.Second).BuildServerInterceptor()),
	)
	go func() {
		log.Println("开始启动grpc服务端...")
		lis, err := net.Listen("tcp", "127.0.0.1:9999")
		if err != nil {
			panic(err)
		}
		// 注册服务
		service.RegisterTestServiceServer(grpcServer, &service.TestService{})
		// 阻塞操作
		err = grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	// 等待grpc服务端启动
	time.Sleep(100 * time.Millisecond)
	log.Println("开始启动grpc客户端...")
	conn, err := grpc.Dial("127.0.0.1:9999", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	grpcClient := pb.NewTestServiceClient(conn)
	t.client = grpcClient
}

func (t *TestSuite) TestLimit() {
	// 服务端的逻辑是睡眠100ms然后返回当前时间戳
	// 测试流程
	// 1. 先开10个goroutine，将获得的时间戳进行排序，相邻的时间戳差不多一致
	// 2. 过了2s内存超过80%，触发限流。也是开10个goroutine，将获得的时间戳进行排序，相邻的时间戳相差100ms左右
	// 3. 5s以后内存下降正常处理请求。10个goroutine获得的时间戳差不多一致


	// endTime 限流结束的时间
	endTime := time.Now().Add(5*time.Second + 10*time.Millisecond)
	timeList1, err := t.getTimeList()
	require.NoError(t.T(), err)
	for i := 1; i < len(timeList1); i++ {
		diff := int64(math.Abs(float64(timeList1[i] - timeList1[i-1])))
		assert.Equal(t.T(), true, diff < 10)
	}
	time.Sleep(2 * time.Second)
	// 测试限流，请求是一个接一个，返回的时间戳差不多相差100ms
	timeList2, err := t.getTimeList()
	require.NoError(t.T(), err)
	for i := 1; i < len(timeList2); i++ {
		diff := int64(math.Abs(float64(timeList2[i] - timeList2[i-1])))
		assert.Equal(t.T(), true, diff >= 100 && diff < 110)
	}
	time.Sleep(endTime.Sub(time.Now()))
	// 内存恢复正常，返回的时间戳差不多一致
	timeList3, err := t.getTimeList()
	for i := 1; i < len(timeList3); i++ {
		diff := int64(math.Abs(float64(timeList3[i] - timeList3[i-1])))
		assert.Equal(t.T(), true, diff < 10)
	}
}

func (t *TestSuite) getTimeList() ([]int64, error) {
	var eg errgroup.Group
	mu := &sync.RWMutex{}
	timeList := make([]int64, 0, 10)
	for i := 0; i < 10; i++ {
		eg.Go(func() error {
			resp, err := t.client.Test(context.Background(), &pb.TestRequest{})
			if err != nil {
				return err
			}
			mu.Lock()
			timeList = append(timeList, resp.Timestamp)
			mu.Unlock()
			return nil
		})
	}
	err := eg.Wait()
	require.NoError(t.T(), err)
	sort.Slice(timeList, func(i, j int) bool {
		return timeList[i] < timeList[j]
	})
	return timeList, nil
}

func TestLimit(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

package case11

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	interceptor2 "interview-cases/case11_20/case11/interceptor"
	pb2 "interview-cases/case11_20/case11/pb"
	"interview-cases/case11_20/case11/service"
	"interview-cases/test"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestCase11_(t *testing.T) {

	db := test.InitDB()
	client := test.InitRedis()
	svc := service.NewArticleService(client, db)
	tokenBucket := interceptor2.NewTokenBucket(5, 0) // 最大容量为 5，每秒产生 0 个令牌，方便测试限流
	// 初始化grpc服务端,注册拦截器
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor2.UnaryServerInterceptor(tokenBucket)),
	)

	go func() {
		log.Println("开始启动grpc服务端...")
		lis, err := net.Listen("tcp", "127.0.0.1:9999")
		if err != nil {
			panic(err)
		}

		// 注册服务
		service.RegisterArticleServiceServer(grpcServer, svc)
		// 阻塞操作
		err = grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	// 借助 GORM 来初始化表结构
	err := db.AutoMigrate(&service.Article{})

	require.NoError(t, err)

	log.Println("开始启动grpc客户端...")
	conn, err := grpc.Dial("127.0.0.1:9999", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	grpcClient := pb2.NewArticleServiceClient(conn)

	testCases := []struct {
		name    string
		req     func() *pb2.ListArticlesRequest
		before  func()
		after   func()
		wantRes *pb2.ListArticlesResponse
		wantErr error
	}{
		{
			name: "没有限流,执行先查询redis,没有数据查询到mysql",
			req: func() *pb2.ListArticlesRequest {
				return &pb2.ListArticlesRequest{Author: "mysql"}
			},
			before: func() {
				// 先检测令牌桶是否还有令牌
				tokens := tokenBucket.Tokens()
				if tokens <= 0 {
					// 默认给足了令牌
					tokenBucket.Add(100)
				}

				// 先清空redis数据
				key := "article:redis"
				err := client.Del(context.Background(), key).Err()
				assert.NoError(t, err)

				// 给mysql插入数据
				article := service.Article{
					ID:      1,
					Title:   "mysql",
					Author:  "mysql",
					Content: "没有限流,最终查询了mysql",
				}
				err = db.WithContext(context.Background()).Create(&article).Error

				assert.NoError(t, err)
			},
			after: func() {
				// 清空mysql
				err := db.Exec("TRUNCATE TABLE `articles`").Error
				assert.NoError(t, err)
			},
			wantRes: &pb2.ListArticlesResponse{
				Articles: []*pb2.Article{
					&pb2.Article{
						Id:      1,
						Title:   "mysql",
						Author:  "mysql",
						Content: "没有限流,最终查询了mysql",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "没有限流,执行先查询redis,并且数据从redis返回成功",
			req: func() *pb2.ListArticlesRequest {
				return &pb2.ListArticlesRequest{Author: "redis"}
			},
			before: func() {
				// 先检测令牌桶是否还有令牌
				tokens := tokenBucket.Tokens()
				if tokens <= 0 {
					// 默认给足了令牌
					tokenBucket.Add(100)
				}

				// 先给redis预加载数据
				key := "article:redis"
				list := []*pb2.Article{
					&pb2.Article{
						Id:      1,
						Title:   "redis",
						Author:  "redis",
						Content: "没限流但是只查询redis",
					},
				}
				val, _ := json.Marshal(&list)
				client.Set(context.Background(), key, val, time.Minute*10)

				assert.NoError(t, err)
			},
			after: func() {
				// 清空redis
				key := "article:redis"
				err := client.Del(context.Background(), key).Err()
				assert.NoError(t, err)
			},
			wantRes: &pb2.ListArticlesResponse{
				Articles: []*pb2.Article{
					&pb2.Article{
						Id:      1,
						Title:   "redis",
						Author:  "redis",
						Content: "没限流但是只查询redis",
					},
				},
			},

			wantErr: nil,
		},
		{
			name: "限流,只查询redis,并且数据从redis返回成功",
			req: func() *pb2.ListArticlesRequest {
				return &pb2.ListArticlesRequest{Author: "redis"}
			},
			before: func() {
				// 先消耗掉令牌，模拟限流环境
				tokens := tokenBucket.Tokens()
				if tokens > 0 {
					tokenBucket.Consume(tokens)
				}

				// 先给redis预加载数据
				key := "article:redis"
				list := []*pb2.Article{
					&pb2.Article{
						Id:      1,
						Title:   "redis",
						Author:  "redis",
						Content: "限流但是只查询redis",
					},
				}
				val, _ := json.Marshal(&list)
				client.Set(context.Background(), key, val, time.Minute*10)

				assert.NoError(t, err)
			},
			after: func() {
				// 清空redis
				key := "article:redis"
				err := client.Del(context.Background(), key).Err()
				assert.NoError(t, err)
			},
			wantRes: &pb2.ListArticlesResponse{
				Articles: []*pb2.Article{
					&pb2.Article{
						Id:      1,
						Title:   "redis",
						Author:  "redis",
						Content: "限流但是只查询redis",
					},
				},
			},

			wantErr: nil,
		},
		{
			name: "限流,只查询redis,并且redis数据不存在，返回nil",
			req: func() *pb2.ListArticlesRequest {
				return &pb2.ListArticlesRequest{Author: "redis"}
			},
			before: func() {
				// 先消耗掉令牌，模拟限流环境
				tokens := tokenBucket.Tokens()
				if tokens > 0 {
					tokenBucket.Consume(tokens)
				}
			},
			after: func() {
				// 清空redis
				key := "article:redis"
				err := client.Del(context.Background(), key).Err()
				assert.NoError(t, err)
			},
			wantRes: nil,
			wantErr: status.Errorf(codes.Unknown, "数据不存在redis"),
		},
	}

	// 开始执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.req()
			tc.before()
			resp, err := grpcClient.ListArticles(context.Background(), req)
			if resp == nil {
				//如果返回的不是nil 直接assert.Equal 会出问题
				assert.Equal(t, tc.wantRes, resp)
			} else {
				assert.Equal(t, tc.wantRes.Articles, resp.Articles)
			}
			assert.Equal(t, tc.wantErr, err)
			tc.after()
		})
	}

	// 关闭grpc服务

}

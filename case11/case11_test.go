package case11

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"interview-cases/case11/interceptor"
	"interview-cases/case11/pb"
	ratelimiter "interview-cases/case11/ratelimit"
	"interview-cases/case11/repo"
	"interview-cases/case11/service"
	"log"
	"net"
	"testing"
	"time"
)

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

type Article struct {
	ID      int32
	Title   string
	Author  string
	Content string
}

type Case11TestSuite struct {
	suite.Suite
	db          *gorm.DB
	client      redis.Cmdable
	svc         *service.ArticleService
	grpcServer  *grpc.Server
	tokenBucket *ratelimiter.TokenBucket
	grpcClient  pb.ArticleServiceClient
}

// 测试方式可以很简单，可以先往redis存数据，限流路线判断是否和我预先存的数据是否匹配
func (s *Case11TestSuite) prepareRedis(key string, list []*pb.Article) {

	val, _ := json.Marshal(&list)
	s.client.Set(context.Background(), key, val, time.Minute*10)
}

// 往mysql插入一条数据，然后插入到mysql里面 取出来判断是否走的是mysql
func (s *Case11TestSuite) prepareMysql(article *repo.Article) {
	s.db.WithContext(context.Background()).Model(&Article{}).Create(article)
}

func (s *Case11TestSuite) SetupSuite() {
	s.db = InitDB()
	s.client = repo.InitRedis()
	s.svc = service.NewArticleService(s.client, s.db)
	s.tokenBucket = ratelimiter.NewTokenBucket(5, 0) // 最大容量为 5，每秒产生 0 个令牌，方便测试限流
	// 初始化grpc服务端,注册拦截器
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryServerInterceptor(s.tokenBucket)),
	)

	go func() {
		log.Println("开始启动grpc服务端...")
		lis, err := net.Listen("tcp", "127.0.0.1:9999")
		if err != nil {
			panic(err)
		}

		// 注册服务
		service.RegisterArticleServiceServer(s.grpcServer, s.svc)
		// 阻塞操作
		err = s.grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	// 借助 GORM 来初始化表结构
	err := s.db.AutoMigrate(&Article{})

	require.NoError(s.T(), err)

	time.Sleep(time.Second * 2)

	log.Println("开始启动grpc客户端...")
	conn, err := grpc.Dial("127.0.0.1:9999", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	//defer func(conn *grpc.ClientConn) {
	//	err := conn.Close()
	//	require.NoError(s.T(), err)
	//}(conn)

	s.grpcClient = pb.NewArticleServiceClient(conn)
}

// 测试没限流 redis->mysql
func (s *Case11TestSuite) TestRedisToMysql() {
	// 先检测令牌桶是否还有令牌
	tokens := s.tokenBucket.Tokens()
	if tokens <= 0 {
		// 默认给足了令牌
		s.tokenBucket.Add(100)
	}

	// 先清空redis数据
	req := &pb.ListArticlesRequest{Author: "mysql"}
	key := "article:redis"
	err := s.client.Del(context.Background(), key).Err()
	require.NoError(s.T(), err)

	// 给mysql插入数据
	article := repo.Article{
		ID:      2,
		Title:   "mysql",
		Author:  req.Author,
		Content: "没有限流,最终查询了mysql",
	}

	equal := toProto([]repo.Article{article})

	s.prepareMysql(&article)

	resp, err := s.grpcClient.ListArticles(context.Background(), req)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), equal, resp.Articles)
	log.Println(resp)
}

// 测试限流只读redis
func (s *Case11TestSuite) TestReadOnlyRedis() {
	// 先消耗掉令牌，模拟限流环境
	tokens := s.tokenBucket.Tokens()
	if tokens > 0 {
		s.tokenBucket.Consume(tokens)
	}

	// 构造缓存数据

	req := &pb.ListArticlesRequest{Author: "redis"}

	key := "article:" + req.Author
	list := []*pb.Article{
		&pb.Article{
			Id:      1,
			Title:   "redis",
			Author:  req.Author,
			Content: "限流只查询redis",
		},
	}

	// 添加redis数据
	s.prepareRedis(key, list)

	resp, err := s.grpcClient.ListArticles(context.Background(), req)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), list, resp.Articles)

}

func (s *Case11TestSuite) TearDownSuite() {
	// 如果你不希望测试结束就删除数据，你把这段代码注释掉
	err := s.db.Exec("TRUNCATE TABLE `articles`").Error
	require.NoError(s.T(), err)
}

func TestCase11(t *testing.T) {
	suite.Run(t, new(Case11TestSuite))
}

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/interview_cases"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&Article{})
	if err != nil {
		panic(err)
	}
	return db

}

func toProto(articles []repo.Article) []*pb.Article {
	pa := make([]*pb.Article, 0, len(articles))
	for _, v := range articles {
		pba := &pb.Article{
			Id:      v.ID,
			Title:   v.Title,
			Author:  v.Author,
			Content: v.Content,
		}

		pa = append(pa, pba)
	}

	return pa
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	pb2 "interview-cases/case11_20/case11/pb"
	"time"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Article struct {
	ID      int32
	Title   string
	Author  string
	Content string
}

type ArticleService struct {
	pb2.UnimplementedArticleServiceServer
	Client redis.Cmdable
	DB     *gorm.DB
}

func NewArticleService(client redis.Cmdable, DB *gorm.DB) *ArticleService {
	return &ArticleService{Client: client, DB: DB}
}

func (s *ArticleService) ListArticles(ctx context.Context, req *pb2.ListArticlesRequest) (*pb2.ListArticlesResponse, error) {
	// 不管有没有限流，redis都是必须查询的
	key := "article:" + req.Author
	resp, err := s.getArticleListFromRedis(ctx, key)
	if err == nil {
		return resp, nil
	}

	// 请求被限流
	if rateLimited, ok := ctx.Value("RateLimited").(bool); ok && rateLimited {
		return nil, errors.New("数据不存在redis")
	}

	resp, err = s.getArticleListFromMySQL(ctx, req.Author)
	if err == nil {
		// 回写redis
		s.setArticleListToRedis(ctx, key, resp.Articles)
	}
	return resp, err

}

func (s *ArticleService) getArticleListFromRedis(ctx context.Context, key string) (*pb2.ListArticlesResponse, error) {
	// 从 Redis 获取文章列表
	res, err := s.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var articles []*pb2.Article
	er := json.Unmarshal(res, &articles)

	if er != nil {
		return nil, er
	}

	return &pb2.ListArticlesResponse{
		Articles: articles,
	}, nil
}

func (s *ArticleService) setArticleListToRedis(ctx context.Context, key string, val []*pb2.Article) error {
	value, _ := json.Marshal(val)
	return s.Client.Set(ctx, key, value, time.Minute*10).Err()

}

func (s *ArticleService) getArticleListFromMySQL(ctx context.Context, author string) (*pb2.ListArticlesResponse, error) {
	// 从 MySQL 获取文章列表
	// 假设这里返回了一个空的响应
	var articles []Article
	err := s.DB.WithContext(ctx).Model(Article{}).Where("author = ?", author).Find(&articles).Error
	if err != nil {
		return nil, err
	}

	return &pb2.ListArticlesResponse{
		Articles: toProto(articles),
	}, nil
}

func toProto(articles []Article) []*pb2.Article {
	pa := make([]*pb2.Article, 0, len(articles))
	for _, v := range articles {
		pba := &pb2.Article{
			Id:      v.ID,
			Title:   v.Title,
			Author:  v.Author,
			Content: v.Content,
		}

		pa = append(pa, pba)
	}

	return pa
}

func RegisterArticleServiceServer(s *grpc.Server, svc *ArticleService) {
	pb2.RegisterArticleServiceServer(s, svc)
}

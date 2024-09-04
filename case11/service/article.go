package service

import (
	"context"
	"encoding/json"
	"interview-cases/case11/pb"
	"interview-cases/case11/repo"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type ArticleService struct {
	pb.UnimplementedArticleServiceServer
	Client redis.Cmdable
	DB     *gorm.DB
}

func NewArticleService(client redis.Cmdable, DB *gorm.DB) *ArticleService {
	return &ArticleService{Client: client, DB: DB}
}

func (s *ArticleService) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (*pb.ListArticlesResponse, error) {
	key := "article:" + req.Author

	// 请求被限流
	if rateLimited, ok := ctx.Value("RateLimited").(bool); ok && rateLimited {
		log.Println("请求限流，只允许从redis进行查询")
		return s.getArticleListFromRedis(ctx, key)
	} else {
		log.Println("请求没有限流，可以访问Mysql进行查询")
		resp, err := s.getArticleListFromRedis(ctx, key)
		if err == nil {
			return resp, nil
		}

		// Redis 中没有数据，访问 MySQL
		resp, err = s.getArticleListFromMySQL(ctx, req.Author)

		if err == nil {
			// 回写redis
			s.setArticleListToRedis(ctx, key, resp.Articles)
		}

		return resp, err
	}

}

func (s *ArticleService) getArticleListFromRedis(ctx context.Context, key string) (*pb.ListArticlesResponse, error) {
	// 从 Redis 获取文章列表
	res, err := s.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var articles []*pb.Article
	er := json.Unmarshal(res, &articles)

	if er != nil {
		return nil, er
	}

	return &pb.ListArticlesResponse{
		Articles: articles,
	}, nil
}

func (s *ArticleService) setArticleListToRedis(ctx context.Context, key string, val []*pb.Article) error {
	value, _ := json.Marshal(val)
	return s.Client.Set(ctx, key, value, time.Minute*10).Err()

}

func (s *ArticleService) getArticleListFromMySQL(ctx context.Context, author string) (*pb.ListArticlesResponse, error) {
	// 从 MySQL 获取文章列表
	// 假设这里返回了一个空的响应
	var articles []repo.Article
	err := s.DB.WithContext(ctx).Model(repo.Article{}).Where("author = ?", author).Find(&articles).Error
	if err != nil {
		return nil, err
	}

	return &pb.ListArticlesResponse{
		Articles: toProto(articles),
	}, nil
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

func RegisterArticleServiceServer(s *grpc.Server, svc *ArticleService) {
	pb.RegisterArticleServiceServer(s, svc)
}

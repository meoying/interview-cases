package cronjob

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	articleDomain "interview-cases/case21_30/case21/cronjob/tri/domain"
	"interview-cases/case21_30/case21/cronjob/tri/service"
	"interview-cases/case21_30/case21/domain"
)

// TriSvc 第三方服务用于模拟前1000名的数据
type TriSvc interface {
	TopN(ctx context.Context, n int) ([]domain.RankItem, error)
}

// 文章服务模拟文章的点赞数据
type articleSvc struct {
	svc service.ArticleSvc
}

func NewTriSvc(svc service.ArticleSvc) TriSvc {
	return &articleSvc{svc: svc}
}

func (t *articleSvc) TopN(ctx context.Context, n int) ([]domain.RankItem, error) {
	articles, err := t.svc.TopN(ctx, n)
	if err != nil {
		return nil, err
	}
	ans := slice.Map(articles, func(idx int, src articleDomain.Article) domain.RankItem {
		return domain.RankItem{
			ID:    src.ID,
			Score: int(src.LikeCnt),
		}
	})
	return ans, nil
}

package cronjob

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"interview-cases/case21_30/case21/cronjob/dao"
	"interview-cases/case21_30/case21/domain"
)

type ArticleSvc interface {
	// BatchCreate 批量创建统计数据
	BatchCreate(ctx context.Context, articles []domain.Article) error
	// TopN 获取前1000名
	TopN(ctx context.Context, n int) ([]domain.Article, error)
}

type articleSvc struct {
	articleDao dao.ArticleStaticDAO
}

func NewArticleSvc(articleDao dao.ArticleStaticDAO) ArticleSvc {
	return &articleSvc{
		articleDao: articleDao,
	}
}

func (a *articleSvc) BatchCreate(ctx context.Context, articles []domain.Article) error {
	articleEntities := slice.Map(articles, func(idx int, src domain.Article) dao.ArticleStatic {
		return dao.ArticleStatic{
			ArticleID: src.ID,
			LikeCnt:   int32(src.LikeCnt),
		}
	})
	return a.articleDao.BatchCreate(ctx, articleEntities)
}

func (a *articleSvc) TopN(ctx context.Context, n int) ([]domain.Article, error) {
	articles, err := a.articleDao.TopN(ctx, n)
	if err != nil {
		return nil, err
	}
	list := slice.Map(articles, func(idx int, src dao.ArticleStatic) domain.Article {
		return domain.Article{
			ID:      src.ArticleID,
			LikeCnt: int(src.LikeCnt),
		}
	})
	return list, nil
}

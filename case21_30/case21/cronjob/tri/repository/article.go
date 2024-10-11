package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"interview-cases/case21_30/case21/cronjob/tri/domain"
	"interview-cases/case21_30/case21/cronjob/tri/repository/dao"
)

type ArticleRepo interface {
	// BatchCreate 批量创建
	BatchCreate(ctx context.Context, articles []domain.Article) error
	// TopN 获取前1000名
	TopN(ctx context.Context,n int) ([]domain.Article, error)
}

type articleRepo struct {
	articleDao dao.ArticleStaticDAO
}

func NewArticleRepo(articleDao dao.ArticleStaticDAO )ArticleRepo {
	return &articleRepo{
		articleDao: articleDao,
	}
}

func (a *articleRepo) BatchCreate(ctx context.Context, articles []domain.Article) error {
	articleEntities := slice.Map(articles, func(idx int, src domain.Article) dao.ArticleStatic {
		return dao.ArticleStatic{
			ArticleID: src.ID,
			LikeCnt:   src.LikeCnt,
		}
	})
	return a.articleDao.BatchCreate(ctx,articleEntities)

}

func (a *articleRepo) TopN(ctx context.Context, n int) ([]domain.Article, error) {
	articles,err := a.articleDao.TopN(ctx,n)
	if err != nil {
		return nil,err
	}
	ans := slice.Map(articles, func(idx int, src dao.ArticleStatic) domain.Article {
		return domain.Article{
			ID: src.ArticleID,
			LikeCnt: src.LikeCnt,
		}
	})
	return ans,nil
}

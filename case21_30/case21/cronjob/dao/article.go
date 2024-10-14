package dao

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
)

// 文章的统计表
type ArticleStatic struct {
	ID        int64 `gorm:"primaryKey"`
	ArticleID int64
	LikeCnt   int32
	// 其他字段就不展示了
}

type ArticleStaticDAO interface {
	// BatchCreate 批量创建
	BatchCreate(ctx context.Context, articles []ArticleStatic) error
	// TopN 获取前1000名
	TopN(ctx context.Context, n int) ([]ArticleStatic, error)
}

type articleStaticDAO struct {
	db   *gorm.DB
	node *snowflake.Node
}

func NewArticleStaticDAO(db *gorm.DB) ArticleStaticDAO {
	node, _ := snowflake.NewNode(0)
	return &articleStaticDAO{
		db:   db,
		node: node,
	}
}

// BatchCreate 使用分库分表按article_id和2取余，结果为0放在article_static_tab0 结果为1放在article_static_tab1
func (a *articleStaticDAO) BatchCreate(ctx context.Context, articles []ArticleStatic) error {
	articleMap := make(map[string][]ArticleStatic)
	for _, article := range articles {
		article.ID = a.node.Generate().Int64()
		tab := fmt.Sprintf("article_static_tab%d", article.ArticleID%2)
		articleList, ok := articleMap[tab]
		if !ok {
			articleList = []ArticleStatic{
				article,
			}
		} else {
			articleList = append(articleList, article)
		}
		articleMap[tab] = articleList
	}
	for tab, as := range articleMap {
		err := a.db.Table(tab).
			Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{
					"like_cnt",
				}),
				Columns: []clause.Column{
					{
						Name: "article_id",
					},
				},
			}).
			Create(as).Error
		if err != nil {
			return err
		}
	}
	return nil

}

func (a *articleStaticDAO) TopN(ctx context.Context, n int) ([]ArticleStatic, error) {
	tabs := []string{
		"article_static_tab0",
		"article_static_tab1",
	}
	mu := &sync.RWMutex{}
	sList := make([][]ArticleStatic, 0, 2)
	// 从每个表读取数据
	var eg errgroup.Group
	for _, tab := range tabs {
		tab := tab
		eg.Go(func() error {
			var staticList []ArticleStatic
			err := a.db.Table(tab).Order("like_cnt desc").Limit(n).Find(&staticList).Error
			if err != nil {
				return err
			}
			mu.Lock()
			sList = append(sList, staticList)
			mu.Unlock()
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	// 归并排序
	return GetSortList(sList, n), nil
}

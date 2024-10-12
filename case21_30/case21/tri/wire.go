//go:build wireinject

package tri

import (
	"github.com/google/wire"
	"gorm.io/gorm"
	"interview-cases/case21_30/case21/tri/repository"
	"interview-cases/case21_30/case21/tri/repository/dao"
	"interview-cases/case21_30/case21/tri/service"
)

func InitModule(db *gorm.DB) (*Module, error) {
	wire.Build(
		dao.NewArticleStaticDAO,
		repository.NewArticleRepo,
		service.NewArticleSvc,
		wire.Struct(new(Module), "*"),
	)
	return new(Module), nil
}

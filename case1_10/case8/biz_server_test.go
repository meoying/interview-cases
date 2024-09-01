package case8

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"interview-cases/test"
	"log/slog"
	"net/http"
)

// StartServer 模拟业务服务器
func StartServer(addr string) {
	r := gin.Default()
	db := InitDb()

	hdl := &BizHandler{
		db: db,
	}
	hdl.RegisterRouter(r)
	r.Run(addr)
}

type BizHandler struct {
	count int64
	db    *gorm.DB
}

func (t *BizHandler) RegisterRouter(server *gin.Engine) {
	server.POST("/handle", func(c *gin.Context) {
		var u UserCase8
		// 拿到 UserCase8 的数据
		if err := c.Bind(&u); err != nil {
			c.String(http.StatusBadRequest, "参数错误")
			slog.Error("参数错误", slog.Any("err", err))
			return
		}
		// 我们这里可以简单模拟一下，真实的业务场景不会那么简单
		err := t.db.Create(&u).Error
		if err != nil {
			c.String(http.StatusInternalServerError, "系统错误")
			slog.Error("系统错误", slog.Any("err", err))
			return
		}
		slog.Info("处理成功", slog.Int64("id", u.ID))
		c.String(http.StatusOK, "OK")
	})
}

type UserCase8 struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	Name      string
	Email     string
	Password  string
	CreatedAt int64
	UpdatedAt int64
}

func InitDb() *gorm.DB {
	db := test.InitDB()
	err := db.AutoMigrate(&UserCase8{})
	if err != nil {
		panic(err)
	}
	return db
}

package case9

import (
	"encoding/json"
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
	// 单个接口
	server.POST("/single", func(c *gin.Context) {
		var u UserCase9
		// 拿到 UserCase9 的数据
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

	server.POST("/batch", func(c *gin.Context) {
		var vals []string
		if err := c.Bind(&vals); err != nil {
			c.String(http.StatusBadRequest, "参数错误")
			slog.Error("参数错误", slog.Any("err", err))
			return
		}

		users := make([]UserCase9, 0, len(vals))
		for _, val := range vals {
			var u UserCase9
			err1 := json.Unmarshal([]byte(val), &u)
			if err1 != nil {
				c.String(http.StatusBadRequest, "参数错误")
				slog.Error("参数错误",
					slog.String("data", val),
					slog.Any("err", err1))
				return
			}
			users = append(users, u)
		}
		// 一次性插入到数据库中。在实践中，批量插入远比单个插入性能要好
		err := t.db.Create(&users).Error
		if err != nil {
			c.String(http.StatusInternalServerError, "系统错误")
			slog.Error("系统错误", slog.Any("err", err))
			return
		}
		slog.Info("处理成功", slog.Int("size", len(users)))
		c.String(http.StatusOK, "OK")
	})
}

type UserCase9 struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	Name      string
	Email     string
	Password  string
	CreatedAt int64
	UpdatedAt int64
}

func InitDb() *gorm.DB {
	db := test.InitDB()
	err := db.AutoMigrate(&UserCase9{})
	if err != nil {
		panic(err)
	}
	return db
}

package index

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"interview-cases/ioc"
	"math/rand"
	"testing"
	"time"
)

func TestCoverIndex(t *testing.T) {
	db := ioc.InitDB()
	err := db.Exec("TRUNCATE TABLE `user`").Error
	require.NoError(t, err)
	for i := 0; i < 100000; i++ {
		user := User{
			Username: generateRandomString(40),
			Email:    generateRandomEmail(),
			Age:      rand.Intn(100),
		}
		err := db.Create(&user).Error
		if err != nil {
			panic("创建数据失败")
		}
	}
}

type User struct {
	ID       uint `gorm:"primaryKey"`
	Username string
	Email    string
	Age      int
}

func (User) TableName() string {
	return "user"
}

// 生成指定长度的随机字符串
func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 生成随机邮箱
func generateRandomEmail() string {
	return fmt.Sprintf("%s@example.com", generateRandomString(77))
}

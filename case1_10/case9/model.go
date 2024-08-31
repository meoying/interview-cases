package case9

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

type User struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func InitDb()*gorm.DB {
	dsn := "root:root@tcp(localhost:13306)/interview_cases?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=1s&readTimeout=3s&writeTimeout=3s"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	initUser(db)
	return db
}

func initUser(db *gorm.DB) {
	db.AutoMigrate(&User{})
}

type UserDAO struct {
	db *gorm.DB
}

// 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	s := make([]byte, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// 生成随机email
func randomEmail() string {
	domains := []string{"example.com", "test.com", "sample.org"}
	return fmt.Sprintf("%s@%s", randomString(8), domains[rand.Intn(len(domains))])
}

// 随机生成几个user
func (u *UserDAO) Insert(number int) error {
	uList := make([]User, 0, number)
	for i := 0; i < number; i++ {
		uList = append(uList, User{
			Name:      randomString(10), // 生成随机名字
			Email:     randomEmail(),    // 生成随机email
			Password:  randomString(12), // 生成随机密码
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
	return u.db.Model(&User{}).Create(&uList).Error
}

package index

import (
	"github.com/stretchr/testify/require"
	"interview-cases/ioc"
	"math/rand"
	"testing"
)

func TestOrderIndex(t *testing.T) {
	db := ioc.InitDB()
	err := db.Exec("TRUNCATE TABLE `products`").Error
	require.NoError(t, err)
	for i := 0; i < 100000; i++ {
		p := Product{
			Name:     generateRandomString(55),
			Price:    rand.Float64(),
			Category: generateRandomString(20),
		}
		err := db.Create(&p).Error
		if err != nil {
			panic("创建数据失败")
		}
	}

}

type Product struct {
	ID       uint `gorm:"primaryKey"`
	Price    float64
	Name     string
	Category string
}

func (Product) TableName() string {
	return "products"
}

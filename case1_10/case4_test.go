package case1_10

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"interview-cases/test"
	"math/rand"
	"testing"
	"time"
)

// Case4TestSuite 模拟利用索引来排序
type Case4TestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (s *Case4TestSuite) SetupSuite() {
	s.db = test.InitDB()
	// 借助 GORM 来初始化表结构
	err := s.db.AutoMigrate(&Order{}, &OrderV1{})
	require.NoError(s.T(), err)
}

func (s *Case4TestSuite) TearDownSuite() {
	// 如果你不希望测试结束就删除数据，你把这段代码注释掉
	err := s.db.Exec("TRUNCATE TABLE `orders`").Error
	require.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `orders_v1`").Error
	require.NoError(s.T(), err)
}

func (s *Case4TestSuite) TestOrderByWithIndex() {
	// 准备数据，你可以调整里面参数来控制数据量
	s.prepareData()
	// 执行查询，为了方便你查看，我这里直接手写 SQL
	start := time.Now()
	// 查询买家为 11 的数据，模拟用户查询数据，并且翻页
	s.iterateOrder()
	duration := time.Since(start)
	s.T().Log("查询 orders 耗时（ms）:", duration.Milliseconds())

	start = time.Now()
	// 查询买家为 11 的数据，模拟用户查询数据，并且翻页
	s.iterateOrderV1()
	duration = time.Since(start)
	s.T().Log("查询 orders_v1 耗时（ms）:", duration.Milliseconds())
}

func (s *Case4TestSuite) iterateOrder() {
	offset, limit := 0, 10
	for i := 0; i < 100; i++ {
		var res []Order
		err := s.db.Where("buyer=?", 11).
			Find(&res).Limit(limit).Offset(offset).Error
		assert.NoError(s.T(), err)
		if len(res) < limit {
			return
		}
		offset += limit
	}
}

func (s *Case4TestSuite) iterateOrderV1() {
	offset, limit := 0, 10
	for i := 0; i < 100; i++ {
		var res []OrderV1
		err := s.db.Where("buyer=?", 11).
			Find(&res).Limit(limit).Offset(offset).Error
		assert.NoError(s.T(), err)
		if len(res) < limit {
			return
		}
		offset += limit
	}
}

func (s *Case4TestSuite) prepareData() {
	// 首先准备数据。这里你可以调整批次大小以及总批次来调整数据库中的数据
	const batchSize = 100
	const batch = 1000
	orders := make([]Order, 0, 100)
	orderV1s := make([]OrderV1, 0, 100)
	now := time.Now().UnixMilli()
	for i := 0; i < batch; i++ {
		for j := 0; j < batchSize; j++ {
			idx := i*batchSize + j
			orders = append(orders, Order{
				SN:    fmt.Sprintf("sn_%d", idx),
				Buyer: int64(idx % 100),
				Extra: `
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
`,
				Ctime: now + rand.Int63(),
				Utime: now,
			})

			orderV1s = append(orderV1s, OrderV1{
				SN:    fmt.Sprintf("sn_%d", idx),
				Buyer: int64(idx % 100),
				Extra: `
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
`,
				Ctime: now + int64(idx),
				Utime: now + int64(idx),
			})
		}

		err := s.db.Create(&orders).Error
		require.NoError(s.T(), err)
		err = s.db.Create(&orderV1s).Error
		require.NoError(s.T(), err)
		// 重置数据，下一轮复用
		orders = orders[:0]
		orderV1s = orderV1s[:0]
	}
}

func TestCase4(t *testing.T) {
	suite.Run(t, new(Case4TestSuite))
}

type Order struct {
	Id int64
	// SN 是订单编号，就是正常你在订单列表看到的那一长串字符串
	SN string

	// 买家 ID，我们同样是在买家 ID 上创建一个索引
	Buyer int64 `gorm:"index"`

	// 模拟别的字段，占据了空间
	Extra string

	// 毫秒时间戳
	// 假设说我们现在的主要业务场景是根据买家 ID 查询订单，然后按照下单时间倒序排序
	// Ctime 上创建了一个索引，独立的索引
	Ctime int64 `gorm:"index"`
	Utime int64
}

func (o Order) TableName() string {
	return "orders"
}

type OrderV1 struct {
	Id int64
	// SN 是订单编号，就是正常你在订单列表看到的那一长串字符串
	SN string

	// 模拟别的字段，占据了空间
	Extra string

	// 买家 ID
	// 我们在 Buyer 和 Ctime 上创建一个联合索引
	// 并且我们的索引列的顺序是（buyer，ctime）
	Buyer int64 `gorm:"index:buyer_ctime"`
	Ctime int64 `gorm:"index:buyer_ctime"`

	Utime int64
}

func (o OrderV1) TableName() string {
	return "orders_v1"
}

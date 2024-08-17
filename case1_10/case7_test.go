// Copyright 2023 ecodeclub
//
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

type Case7TestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (s *Case7TestSuite) SetupSuite() {
	s.db = test.InitDB()
	// 借助 GORM 来初始化表结构
	err := s.db.AutoMigrate(&Order{}, &OrderV1{})
	require.NoError(s.T(), err)
}

func (s *Case7TestSuite) TearDownSuite() {
	// 如果你不希望测试结束就删除数据，你把这段代码注释掉
	//err := s.db.Exec("TRUNCATE TABLE `orders`").Error
	//require.NoError(s.T(), err)
}

func (s *Case7TestSuite) TestPageWithWhere() {
	// 准备数据，你可以调整里面参数来控制数据量
	//s.prepareData()
	// 执行查询，为了方便你查看，我这里直接手写 SQL
	start := time.Now()
	var res []Order
	err := s.db.Where("buyer=?", 1).Limit(10).Offset(10000).Find(&res).Error
	duration := time.Since(start)
	require.NoError(s.T(), err)
	s.T().Log("深度分页 耗时（ms）:", duration.Milliseconds())

	start = time.Now()
	// 上一页的最大 id，我们就直接用前一个查询查出来的
	lastId := res[0].Id
	res = res[:0]
	err = s.db.Where("buyer=? AND id >= ?", 1, lastId).
		Limit(10).Offset(0).Find(&res).Error
	duration = time.Since(start)
	require.NoError(s.T(), err)
	s.T().Log("深度分页优化后 耗时（ms）:", duration.Milliseconds())

}

func (s *Case7TestSuite) iterateOrder() {
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

func (s *Case7TestSuite) prepareData() {
	// 首先准备数据。这里你可以调整批次大小以及总批次来调整数据库中的数据
	const batchSize = 100
	const batch = 1000
	orders := make([]Order, 0, 100)
	now := time.Now().UnixMilli()
	for i := 0; i < batch; i++ {
		for j := 0; j < batchSize; j++ {
			idx := i*batchSize + j
			orders = append(orders, Order{
				SN:    fmt.Sprintf("sn_%d", idx),
				Buyer: int64(idx % 10),
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

		}

		err := s.db.Create(&orders).Error
		require.NoError(s.T(), err)
		// 重置数据，下一轮复用
		orders = orders[:0]
	}
}

func TestCase7(t *testing.T) {
	suite.Run(t, new(Case7TestSuite))
}

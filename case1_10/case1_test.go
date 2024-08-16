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
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"interview-cases/test"
	"time"
)

type Case1TestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (s *Case1TestSuite) SetupSuite() {
	s.db = test.InitDB()
	err := s.db.AutoMigrate(&UserBefore{}, &UserAfter{})
	require.NoError(s.T(), err)
	// 插入数据的时候打印日志很麻烦，所以这里调高了日志级别
	//s.db.Logger = logger.New(log.Default(), logger.Config{
	//	LogLevel: logger.Error,
	//})

	err = s.db.Exec("TRUNCATE table `users_before`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE table `users_after`").Error
	assert.NoError(s.T(), err)
}

func (s *Case1TestSuite) TearDownSuite() {
	// 如果你要清除数据，可以直接删掉整个表
	//err := s.db.Exec("DROP table `users_before`").Error
	//assert.NoError(s.T(), err)
	//err = s.db.Exec("DROP table `users_after`").Error
	//assert.NoError(s.T(), err)
}

// 测试覆盖索引的效力
func (s *Case1TestSuite) TestCoverIndex() {
	// 超时，如果你的机器性能不好，可以调大
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	db := s.db.WithContext(ctx)
	// 生成数据
	// 你可以调整这个参数来控制生成的数据数量
	// 当然如果要是很多，你的内存可能爆掉，你自己评估
	const (
		// 多少批
		num = 100
		// 每一批大小
		batchSize = 100
	)

	now := time.Now().UnixMilli()

	for i := 0; i < num; i++ {
		befores := make([]UserBefore, 0, batchSize)
		afters := make([]UserAfter, 0, batchSize)
		for j := 0; j < batchSize; j++ {
			id := uint(i*100+j) + 1
			befores = append(befores, UserBefore{
				ID:       id,
				Username: fmt.Sprintf("username_%d", id),
				Password: fmt.Sprintf("password_%d", id),
				Age:      18,
				Avatar:   fmt.Sprintf("头像_%d", id),
				Ctime:    now,
				Utime:    now,
			})
			afters = append(afters, UserAfter{
				ID:       id,
				Username: fmt.Sprintf("username_%d", id),
				Password: fmt.Sprintf("password_%d", id),
				Age:      35,
				Avatar:   fmt.Sprintf("头像_%d", id),
				Ctime:    now,
				Utime:    now,
			})
		}

		err := db.Create(&befores).Error
		require.NoError(s.T(), err)
		err = db.Create(&afters).Error
		require.NoError(s.T(), err)
	}

	// 现在就是来对比查询时间了
	// 而后在控制台分别执行下面两个语句:
	// SELECT * FROM users_before WHERE username='username_999';
	// SELECT username, password FROM users_after WHERE username='username_999';
}

// UserBefore 改造前
type UserBefore struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique;type:varchar(255)"`
	Password string `gorm:"type:varchar(255)"`
	Age      int
	// 头像
	Avatar string `gorm:"type:varchar(255)"`

	// 其它字段，一般来说，额外字段越多，改进的效果就越好

	// 时间戳
	Utime int64
	Ctime int64
}

func (UserBefore) TableName() string {
	return "users_before"
}

// UserAfter 改造后
type UserAfter struct {
	ID uint `gorm:"primaryKey"`
	// 正常这需要唯一索引确保不冲突，也需要联合索引来加速查询
	Username string `gorm:"index:name_pwd;type:varchar(255)"`
	Password string `gorm:"index:name_pwd;type:varchar(255)"`
	Age      int
	Avatar   string `gorm:"type:varchar(255)"`
	Utime    int64
	Ctime    int64
}

func (UserAfter) TableName() string {
	return "users_after"
}

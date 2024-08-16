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

import "github.com/stretchr/testify/suite"

type Case5TestSuite struct {
	suite.Suite
}

// 在当前目录下执行  go build -gcflags '-m'  就可以看到逃逸分析的输出
// 返回指针与结构体
func (s *Case5TestSuite) TestReturnPointer() {
	GetUserV1(123)
	GetUserV2(1234)
}

// 测试 channel
func (s *Case5TestSuite) TestArraySlice() {
	ch1 := make(chan User, 2)
	ch2 := make(chan *User, 2)

	ch1 <- User{Name: "Tom"}
	ch2 <- &User{Name: "Jerry"}
}

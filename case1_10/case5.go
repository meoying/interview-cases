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

func GetUserV1(id int64) *User {
	return &User{
		Id:   id,
		Name: "Tom",
	}
}

func GetUserV2(id int64) User {
	return User{
		Id:   id,
		Name: "Tom",
	}
}

type User struct {
	Id   int64
	Name string
}

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

package case17

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"interview-cases/test"
	"testing"
	"time"
)

func TestLimiter_Allow(t *testing.T) {
	rdb := test.InitRedis()
	key := "case17/limiter_base_list"
	limiter := NewLimiter(rdb, time.Minute,
		2, key)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	defer func() {
		err := rdb.Del(ctx, key).Err()
		if err != nil {
			t.Log("清理数据失败", err)
		}
	}()
	ok, err := limiter.Allow(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = limiter.Allow(ctx)
	require.NoError(t, err)
	assert.True(t, ok)

	// 这一次就会拒绝
	ok, err = limiter.Allow(ctx)
	require.NoError(t, err)
	assert.False(t, ok)
}

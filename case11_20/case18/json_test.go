package case18

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"interview-cases/test"
	"testing"
	"time"
)

func TestJson(t *testing.T) {
	// 初始化
	rdb := test.InitRedis()
	jsonClient := NewJsonRedis(rdb)
	// 设置键值
	err := jsonClient.SetBusiness()
	require.NoError(t, err)

	startTime := time.Now().UnixMilli()
	for i := 0; i < 1000; i++ {
		res, err := jsonClient.MyBusiness()
		require.NoError(t, err)
		assert.Equal(t, "NVIDIA GTX 1650", res)
	}
	endTime := time.Now().UnixMilli()
	fmt.Printf("通过直接存储json花了 %dms\n", endTime-startTime)

	startTime = time.Now().UnixMilli()
	for i := 0; i < 1000; i++ {
		res, err := jsonClient.MyBusinessV1()
		require.NoError(t, err)
		assert.Equal(t, "NVIDIA GTX 1650", res)
	}
	endTime = time.Now().UnixMilli()
	fmt.Printf("优化成map花了 %dms\n", endTime-startTime)

}

package case18

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed test.json
	testJson string
)

type JsonRedis struct {
	client redis.Cmdable
}

func NewJsonRedis(client redis.Cmdable) *JsonRedis {
	return &JsonRedis{
		client: client,
	}
}

// 设置key
func (j *JsonRedis) SetBusiness() error {
	// 直接以json存储
	err := j.client.Set(context.Background(), "myBusiness", testJson, 0).Err()
	if err != nil {
		return err
	}
	// 优化成map存储
	kvmap := make(map[string]any, 32)
	err = json.Unmarshal([]byte(testJson), &kvmap)
	if err != nil {
		return err
	}
	return j.client.HMSet(context.Background(), "myBusinessV1", kvmap).Err()
}

// 直接获取json
func (j *JsonRedis) MyBusiness() (string, error) {
	res, err := j.client.Get(context.Background(), "myBusiness").Result()
	if err != nil {
		return "", err
	}
	kvmap := make(map[string]any, 32)
	err = json.Unmarshal([]byte(res), &kvmap)
	return kvmap["product_2_graphics"].(string), nil
}

// 优化成map获取

func (j *JsonRedis) MyBusinessV1() (string, error) {
	res,err := j.client.HMGet(context.Background(), "myBusinessV1", "product_2_graphics").Result()
	if err != nil {
		return "",err
	}
	return res[0].(string),nil
}

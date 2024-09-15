package case_648

import (
	"fmt"
	"io"
	"os"
)

type Node struct {
	name    string
	address string
}

func (n *Node) GetCache(uid int) (string, error) {
	fileName := fmt.Sprintf("cache/c_%s_%d.txt", n.address, uid)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	if len(content) == 0 {
		cache := fmt.Sprintf("节点%s缓存", n.name)
		_, err = file.WriteString(cache)
		if err != nil {
			return "", err
		}
		content = []byte(cache)
	}

	return string(content), nil
}

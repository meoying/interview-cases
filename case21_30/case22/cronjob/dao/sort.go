package dao

import (
	"github.com/ecodeclub/ekit/queue"
)

type item struct {
	// 下标
	index int
	// 当前数据
	curData ArticleStatic
	// 队列编号
	number int
}

func itemComparator(src item, dst item) int {
	if src.curData.LikeCnt > dst.curData.LikeCnt {
		return -1
	} else if src.curData.LikeCnt  == dst.curData.LikeCnt  {
		return 0
	} else {
		return 1
	}

}

// GetSortList 从多个排完序的数组中，找出全部数据的前n个数据
func GetSortList(dataLists [][]ArticleStatic, n int) []ArticleStatic {
	// 初始化优先级队列
	itemQueue := queue.NewPriorityQueue[item](0, itemComparator)
	for idx, dataList := range dataLists {
		if len(dataList) == 0 {
			continue
		}
		_ = itemQueue.Enqueue(item{
			curData: dataList[0],
			number:  idx,
		})
	}
	ans := make([]ArticleStatic, 0, n)
	for i := 0; i < n; i++ {
		v, err := itemQueue.Dequeue()
		// 如果队列为空就返回
		if err != nil {
			return ans
		}
		ans = append(ans, v.curData)
		// 将下一个元素放入队列
		nextIndex := v.index + 1
		if nextIndex > len(dataLists[v.number])-1 {
			continue
		}
		v.index = nextIndex
		v.curData = dataLists[v.number][nextIndex]
		_ = itemQueue.Enqueue(v)
	}
	return ans

}

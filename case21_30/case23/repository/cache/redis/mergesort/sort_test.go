package mergesort

import (
	"github.com/stretchr/testify/assert"
	"interview-cases/case21_30/case23/domain"
	"testing"
)

func TestSort(t *testing.T) {
	testcases := []struct {
		name     string
		dataList [][]domain.RankItem
		n        int
		wantData []domain.RankItem
	}{
		{
			name: "元素足够",
			dataList: [][]domain.RankItem{
				{
					{
						ID:    5,
						Score: 5,
					},
					{
						ID:    3,
						Score: 3,
					},
				},
				{
					{
						ID:    10,
						Score: 10,
					},
					{
						ID:    9,
						Score: 9,
					},
					{
						ID:    6,
						Score: 6,
					},
					{
						ID:    2,
						Score: 2,
					},
				},
				{
					{
						ID:    8,
						Score: 8,
					},
					{
						ID:    7,
						Score: 7,
					},
					{
						ID:    4,
						Score: 4,
					},
				},
				{
					{
						ID:    1,
						Score: 1,
					},
				},
			},
			n: 5,
			wantData: []domain.RankItem{
				{
					ID:    10,
					Score: 10,
				},
				{
					ID:    9,
					Score: 9,
				},
				{
					ID:    8,
					Score: 8,
				},
				{
					ID:    7,
					Score: 7,
				},
				{
					ID:    6,
					Score: 6,
				},
			},
		},
		{
			name: "元素不足",
			dataList: [][]domain.RankItem{
				{
					{
						ID:    5,
						Score: 5,
					},
					{
						ID:    3,
						Score: 3,
					},
				},
				{
					{
						ID:    10,
						Score: 10,
					},
					{
						ID:    9,
						Score: 9,
					},
					{
						ID:    6,
						Score: 6,
					},
					{
						ID:    2,
						Score: 2,
					},
				},
				{
					{
						ID:    8,
						Score: 8,
					},
					{
						ID:    7,
						Score: 7,
					},
					{
						ID:    4,
						Score: 4,
					},
				},
				{
					{
						ID:    1,
						Score: 1,
					},
				},
			},
			n: 11,
			wantData: []domain.RankItem{
				{
					ID:    10,
					Score: 10,
				},
				{
					ID:    9,
					Score: 9,
				},
				{
					ID:    8,
					Score: 8,
				},
				{
					ID:    7,
					Score: 7,
				},
				{
					ID:    6,
					Score: 6,
				},
				{
					ID:    5,
					Score: 5,
				},
				{
					ID:    4,
					Score: 4,
				},
				{
					ID:    3,
					Score: 3,
				},
				{
					ID:    2,
					Score: 2,
				},
				{
					ID:    1,
					Score: 1,
				},
			},
		},
		{
			name: "其中有一个数组没有元素",
			dataList: [][]domain.RankItem{
				{
					{
						ID:    5,
						Score: 5,
					},
					{
						ID:    3,
						Score: 3,
					},
				},
				{
					{
						ID:    10,
						Score: 10,
					},
					{
						ID:    9,
						Score: 9,
					},
					{
						ID:    6,
						Score: 6,
					},
					{
						ID:    2,
						Score: 2,
					},
				},
				{
					{
						ID:    8,
						Score: 8,
					},
					{
						ID:    7,
						Score: 7,
					},
					{
						ID:    4,
						Score: 4,
					},
				},
			},
			wantData: []domain.RankItem{
				{
					ID:    10,
					Score: 10,
				},
				{
					ID:    9,
					Score: 9,
				},
				{
					ID:    8,
					Score: 8,
				},
				{
					ID:    7,
					Score: 7,
				},
				{
					ID:    6,
					Score: 6,
				},
				{
					ID:    5,
					Score: 5,
				},
			},
			n: 6,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualData := GetSortList(tc.dataList, tc.n)
			assert.Equal(t, tc.wantData, actualData)
		})

	}
}

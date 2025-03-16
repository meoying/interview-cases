package case35

import (
	"interview-cases/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// 测试专用的订单模型
type TestOrder struct {
	ID     uint `gorm:"primaryKey,autoIncrement"`
	Uid    int64
	Status string `gorm:"type:varchar(20)"`
	// 。。。 其他字段
}

func TestCountOptimization(t *testing.T) {
	// 初始化数据库连接
	db := test.InitDB()

	// 准备测试环境
	t.Run("Prepare", func(t *testing.T) {
		// 创建测试表
		err := db.AutoMigrate(&TestOrder{})
		require.NoError(t, err, "创建测试表失败")

		// 清空旧数据
		err = db.Exec("TRUNCATE TABLE test_orders").Error
		require.NoError(t, err, "清空测试表失败")

		// 批量插入测试数据
		start := time.Now()
		err = db.Transaction(func(tx *gorm.DB) error {
			// 使用批量插入优化
			batchSize := 1000
			for i := 0; i < 100000; i += batchSize {
				end := i + batchSize
				if end > 100000 {
					end = 100000
				}
				models := make([]TestOrder, 0, 32)
				for j := i; j < end; j++ {
					status := "pending"
					if j%2 == 0 {
						status = "completed"
					}
					models = append(models, TestOrder{
						Status: status,
						Uid:    23,
					})
				}
				terr := tx.Create(&models).Error
				require.NoError(t, terr)

			}
			return nil
		})
		require.NoError(t, err, "插入测试数据失败")
		t.Logf("插入10万条数据耗时: %v", time.Since(start))

		// 验证数据准确性
		var count int64
		err = db.Table("test_orders").
			Where("status = ?", "completed").
			Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(50000), count, "测试数据不准确")
	})
	// 性能对比测试
	t.Run("PerformanceComparison", func(t *testing.T) {
		var (
			withoutIndexTime time.Duration
			withIndexTime    time.Duration
			count            int64
		)

		// 无索引测试
		t.Run("WithoutIndex", func(t *testing.T) {
			start := time.Now()
			err := db.Model(&TestOrder{}).
				Where("status = ?", "completed").
				Count(&count).Error
			withoutIndexTime = time.Since(start)

			assert.NoError(t, err)
			assert.Equal(t, int64(50000), count)
			t.Logf("无索引查询耗时: %v", withoutIndexTime)
		})

		// 创建索引
		require.NoError(t, db.Exec("CREATE INDEX idx_status ON test_orders(status)").Error)

		// 带索引测试
		t.Run("WithIndex", func(t *testing.T) {
			start := time.Now()
			err := db.Model(&TestOrder{}).
				Where("status = ?", "completed").
				Count(&count).Error
			withIndexTime = time.Since(start)

			assert.NoError(t, err)
			assert.Equal(t, int64(50000), count)
			t.Logf("带索引查询耗时: %v", withIndexTime)
		})

		// 性能提升验证
		t.Logf("原时间 %v, 优化后: %v", withoutIndexTime, withIndexTime)
	})

	// 执行计划验证
	t.Run("ExplainPlan", func(t *testing.T) {
		var explainResult struct {
			QueryPlan string `gorm:"column:EXPLAIN"`
		}

		err := db.Raw(`
			EXPLAIN FORMAT=JSON 
			SELECT COUNT(*) FROM test_orders 
			WHERE status = 'completed'
		`).Scan(&explainResult).Error

		require.NoError(t, err)
		t.Log("执行计划分析:\n", explainResult.QueryPlan)
		assert.Contains(t, explainResult.QueryPlan, "idx_status",
			"执行计划应显示使用idx_status索引")
	})

	// 增强的清理逻辑
	t.Cleanup(func() {
		if t.Failed() {
			t.Log("测试失败，保留测试数据用于排查")
			return
		}

		// 删除索引（如果存在）
		db.Exec("DROP INDEX idx_status ON test_orders")

		// 删除测试表
		err := db.Migrator().DropTable("test_orders")
		if err != nil {
			t.Logf("清理测试表失败: %v", err)
		}
	})
}

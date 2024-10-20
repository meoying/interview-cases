package v2

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	client     *Client
	servers    []*httptest.Server
	serverURLs []string
}

func (suite *IntegrationTestSuite) SetupTest() {
	suite.client = NewClient()
	suite.servers = make([]*httptest.Server, 5)
	suite.serverURLs = make([]string, 5)

	for i := 0; i < 5; i++ {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Query().Get("error") {
			case "network":
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusServiceUnavailable)
			case "timeout":
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusRequestTimeout)
			case "throttle":
				w.WriteHeader(http.StatusTooManyRequests)
			case "circuit_breaker":
				w.WriteHeader(http.StatusServiceUnavailable)
			default:
				w.WriteHeader(http.StatusOK)
			}
		}))
		suite.servers[i] = server
		suite.serverURLs[i] = server.URL
		suite.client.AddNode(server.URL, 50)
	}
}

func (suite *IntegrationTestSuite) TearDownTest() {
	for _, server := range suite.servers {
		server.Close()
	}
}

// 1. 正常请求场景
func (suite *IntegrationTestSuite) TestNormalRequests() {
	for _, url := range suite.serverURLs {
		suite.sendRequestAndCheckWeight(url, "", 50)
	}
}

// 2. 单一错误类型场景
func (suite *IntegrationTestSuite) TestSingleErrorTypes() {
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=network", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=timeout", 49)
	suite.sendRequestAndCheckWeight(suite.serverURLs[2], "error=throttle", 25)
	suite.sendRequestAndCheckWeight(suite.serverURLs[3], "error=circuit_breaker", 0)
}

// 3. 错误恢复场景
func (suite *IntegrationTestSuite) TestErrorRecovery() {
	// 网络异常恢复
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=network", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "", 0)

	// 超时错误恢复
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=timeout", 49)
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "", 49)

	// 限流/降级状态恢复
	suite.sendRequestAndCheckWeight(suite.serverURLs[2], "error=throttle", 25)
	suite.sendRequestAndCheckWeight(suite.serverURLs[2], "", 25)

	// 熔断状态恢复
	suite.sendRequestAndCheckWeight(suite.serverURLs[3], "error=circuit_breaker", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[3], "", 0)
}

// 4. 多节点混合错误场景
func (suite *IntegrationTestSuite) TestMixedErrors() {
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=network", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=timeout", 49)
	suite.sendRequestAndCheckWeight(suite.serverURLs[2], "error=throttle", 25)
	suite.sendRequestAndCheckWeight(suite.serverURLs[3], "", 50)
}

// 5. 连续错误场景
func (suite *IntegrationTestSuite) TestConsecutiveErrors() {
	// 同一节点连续相同错误
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=timeout", 49)
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=timeout", 48)

	// 同一节点连续不同错误
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=throttle", 25)
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=network", 0)
}

// 6. 权重边界测试
func (suite *IntegrationTestSuite) TestWeightBoundaries() {
	t := suite.T()
	t.Skip()
	// 验证最小权重
	suite.client.AddNode("test_min", 1)
	suite.sendRequestAndCheckWeight("test_min", "error=timeout", 1)

	// 验证最大权重
	suite.client.AddNode("test_max", 100)
	suite.sendRequestAndCheckWeight("test_max", "", 100)

	// 验证权重为0的行为
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=network", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "", 0)
}

// 7. 长期运行场景
func (suite *IntegrationTestSuite) TestLongRunning() {
	url := suite.serverURLs[0]
	suite.sendRequestAndCheckWeight(url, "", 50)
	suite.sendRequestAndCheckWeight(url, "error=timeout", 49)
	suite.sendRequestAndCheckWeight(url, "error=throttle", 24)
	suite.sendRequestAndCheckWeight(url, "", 24)
	suite.sendRequestAndCheckWeight(url, "error=network", 0)
	suite.sendRequestAndCheckWeight(url, "", 0)
}

// 8. 并发请求场景
func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			suite.sendRequestAndCheckWeight(suite.serverURLs[index], "error=timeout", 49)
		}(i)
	}
	wg.Wait()

	for _, url := range suite.serverURLs {
		weight, _ := suite.client.GetWeight(url)
		suite.Equal(49, weight)
	}
}

// 9. 节点动态添加/删除场景
// func (suite *IntegrationTestSuite) TestDynamicNodeManagement() {
// 	// 添加新节点
// 	newURL := "http://new-node.com"
// 	suite.client.AddNode(newURL, 50)
// 	weight, exists := suite.client.GetWeight(newURL)
// 	suite.True(exists)
// 	suite.Equal(50, weight)
//
// 	// 删除节点
// 	suite.client.RemoveNode(suite.serverURLs[0])
// 	_, exists = suite.client.GetWeight(suite.serverURLs[0])
// 	suite.False(exists)
// }

// 全部节点不可用场景
func (suite *IntegrationTestSuite) TestAllNodesUnavailable() {
	for _, url := range suite.serverURLs {
		suite.sendRequestAndCheckWeight(url, "error=network", 0)
	}

	// 验证所有节点权重为0
	for _, url := range suite.serverURLs {
		weight, _ := suite.client.GetWeight(url)
		suite.Equal(0, weight)
	}
}

func (suite *IntegrationTestSuite) sendRequestAndCheckWeight(url, queryParam string, expectedWeight int) {
	resp, err := http.Get(url + "?" + queryParam)
	suite.NoError(err)
	defer resp.Body.Close()

	var clientErr error
	switch resp.StatusCode {
	case http.StatusServiceUnavailable:
		if queryParam == "error=network" {
			clientErr = ErrNetworkFailure
		} else {
			clientErr = ErrCircuitBreaker
		}
	case http.StatusRequestTimeout:
		clientErr = ErrTimeout
	case http.StatusTooManyRequests:
		clientErr = ErrThrottling
	}

	suite.client.AdjustWeight(url, clientErr)
	weight, _ := suite.client.GetWeight(url)
	suite.Equal(expectedWeight, weight)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func TestIntegration(t *testing.T) {
	// 启动服务器
	server := NewServer(":8080")
	go server.ListenAndServe()
	defer server.Close()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	client := NewClient()
	serverURL := "http://localhost:8080"
	client.AddNode(serverURL, 50)

	testCases := []struct {
		name           string
		queryParam     string
		expectedWeight int
		expectedError  error
	}{
		{"正常请求", "", 50, nil},
		{"网络错误", "error=network", 0, ErrNetworkFailure},
		{"超时错误", "error=timeout", 49, ErrTimeout},
		{"限流错误", "error=throttle", 25, ErrThrottling},
		{"熔断错误", "error=circuit_breaker", 0, ErrCircuitBreaker},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(serverURL + "/?" + tc.queryParam)
			assert.NoError(t, err, "HTTP请求不应返回错误")
			resp.Body.Close()

			var clientErr error
			switch resp.StatusCode {
			case http.StatusServiceUnavailable:
				if tc.queryParam == "error=network" {
					clientErr = ErrNetworkFailure
				} else {
					clientErr = ErrCircuitBreaker
				}
			case http.StatusRequestTimeout:
				clientErr = ErrTimeout
			case http.StatusTooManyRequests:
				clientErr = ErrThrottling
			}

			client.AdjustWeight(serverURL, clientErr)
			weight, _ := client.GetWeight(serverURL)
			assert.Equal(t, tc.expectedWeight, weight, "节点权重应符合预期")
			assert.Equal(t, tc.expectedError, clientErr, "错误类型应符合预期")

			// 重置权重以准备下一次测试
			client.AddNode(serverURL, 50)
		})
	}
}

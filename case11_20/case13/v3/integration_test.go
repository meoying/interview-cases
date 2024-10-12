package v3

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

type IntegrationTestSuite struct {
	suite.Suite
	client     *Client
	servers    []*httptest.Server
	serverURLs []string
}

func (suite *IntegrationTestSuite) SetupTest() {
	n := 5
	suite.client = NewClient()
	suite.servers = make([]*httptest.Server, n)
	suite.serverURLs = make([]string, n)

	for i := 0; i < n; i++ {
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
		suite.client.AddNode(server.URL, 100)
	}
}

func (suite *IntegrationTestSuite) TearDownTest() {
	for _, server := range suite.servers {
		server.Close()
	}
}

// 正常请求场景
func (suite *IntegrationTestSuite) TestNormalRequests() {
	for _, url := range suite.serverURLs {
		suite.sendRequestAndCheckWeight(url, "", 100)
	}
}

// 单一错误类型场景
func (suite *IntegrationTestSuite) TestSingleErrorTypes() {
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=network", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=timeout", 99)
	suite.sendRequestAndCheckWeight(suite.serverURLs[2], "error=throttle", 50)
	suite.sendRequestAndCheckWeight(suite.serverURLs[3], "error=circuit_breaker", 0)
}

// 3. 错误恢复场景
func (suite *IntegrationTestSuite) TestErrorRecovery() {
	// 网络异常恢复
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=network", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "", 0)

	// 超时错误恢复
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=timeout", 99)
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "", 99)

	// 限流/降级状态恢复
	suite.sendRequestAndCheckWeight(suite.serverURLs[2], "error=throttle", 50)
	suite.sendRequestAndCheckWeight(suite.serverURLs[2], "", 50)

	// 熔断状态恢复
	suite.sendRequestAndCheckWeight(suite.serverURLs[3], "error=circuit_breaker", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[3], "", 0)
}

// 4. 多节点混合错误场景
func (suite *IntegrationTestSuite) TestMixedErrors() {
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=network", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=timeout", 99)
	suite.sendRequestAndCheckWeight(suite.serverURLs[2], "error=throttle", 50)
	suite.sendRequestAndCheckWeight(suite.serverURLs[3], "", 100)
}

// 5. 连续错误场景
func (suite *IntegrationTestSuite) TestConsecutiveErrors() {
	// 同一节点连续相同错误
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=timeout", 99)
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=timeout", 98)

	// 同一节点连续不同错误
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=throttle", 50)
	suite.sendRequestAndCheckWeight(suite.serverURLs[1], "error=network", 0)
}

// 6. 权重边界测试
func (suite *IntegrationTestSuite) TestWeightBoundaries() {
	// 验证最小权重
	minServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestTimeout)
	}))
	defer minServer.Close()
	suite.client.AddNode(minServer.URL, 1)
	suite.sendRequestAndCheckWeight(minServer.URL, "error=timeout", 1)

	// 验证最大权重
	maxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer maxServer.Close()
	suite.client.AddNode(maxServer.URL, 100)
	suite.sendRequestAndCheckWeight(maxServer.URL, "", 100)

	// 验证权重为0的行为
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "error=network", 0)
	suite.sendRequestAndCheckWeight(suite.serverURLs[0], "", 0)
}

// 长期运行场景
func (suite *IntegrationTestSuite) TestLongRunning() {
	url := suite.serverURLs[0]
	suite.sendRequestAndCheckWeight(url, "", 100)
	suite.sendRequestAndCheckWeight(url, "error=timeout", 99)
	suite.sendRequestAndCheckWeight(url, "error=throttle", 49)
	suite.sendRequestAndCheckWeight(url, "", 49)
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
			suite.sendRequestAndCheckWeight(suite.serverURLs[index], "error=timeout", 99)
		}(i)
	}
	wg.Wait()

	for _, url := range suite.serverURLs {
		weight, _ := suite.client.GetWeight(url)
		suite.Equal(99, weight)
	}
}

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

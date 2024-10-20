package v2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerHandleRequest(t *testing.T) {
	server := NewServer(":9527")

	testCases := []struct {
		name           string
		queryParam     string
		expectedStatus int
	}{
		{"正常请求", "", http.StatusOK},
		{"网络错误", "error=network", http.StatusServiceUnavailable},
		{"超时错误", "error=timeout", http.StatusRequestTimeout},
		{"限流错误", "error=throttle", http.StatusTooManyRequests},
		{"熔断错误", "error=circuit_breaker", http.StatusServiceUnavailable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tc.queryParam, nil)
			w := httptest.NewRecorder()

			server.handleRequest(w, req)
			assert.Equal(t, tc.expectedStatus, w.Code, "服务器应返回预期的状态码")
		})
	}
}

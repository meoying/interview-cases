package v2

import (
	"net/http"
	"time"
)

// Server 表示模拟的服务器
type Server struct {
	http.Server
}

// NewServer 创建一个新的服务器实例
func NewServer(addr string) *Server {
	server := &Server{}
	server.Addr = addr
	server.Handler = http.HandlerFunc(server.handleRequest)
	return server
}

// handleRequest 处理incoming请求，模拟各种错误情况
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Query().Get("error") {
	case "network":
		// 模拟网络错误
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusServiceUnavailable)
	case "timeout":
		// 模拟超时
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusRequestTimeout)
	case "throttle":
		// 模拟限流/降级
		w.WriteHeader(http.StatusTooManyRequests)
	case "circuit_breaker":
		// 模拟熔断
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		// 正常响应
		w.WriteHeader(http.StatusOK)
	}
}

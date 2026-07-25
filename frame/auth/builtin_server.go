package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	auth_group "github.com/HonLaderDev/HonLaderAuth-core/group"
	"github.com/gin-gonic/gin"
)

const honLaderAuthRoute = "/api/honlader"

// AuthServer 管理内置 HonLader 验证服务生命周期。
type AuthServer struct {
	mu     sync.Mutex
	url    string
	cancel context.CancelFunc
}

// Start 启动内置验证服务并返回 /api/honlader 地址。
func (s *AuthServer) Start(parent context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.url != "" {
		return s.url + honLaderAuthRoute, nil
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	hla, err := auth_group.HonLaderAuthConfig{}.New()
	if err != nil {
		_ = listener.Close()
		return "", err
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	hla.Register(engine.Group("/"))

	ctx, cancel := context.WithCancel(parent)
	server := &http.Server{Handler: engine}
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("内置验证服务退出：%v\n", err)
		}
	}()

	s.cancel = cancel
	s.url = "http://" + listener.Addr().String()
	return s.url + honLaderAuthRoute, nil
}

// Stop 停止内置验证服务。
func (s *AuthServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	s.cancel = nil
	s.url = ""
}

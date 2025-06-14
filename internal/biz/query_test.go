package biz

import (
	"fmt"
	"testing"
)

// 写一个关于服务器的函数式配置项模式
type Server struct {
	Host string
	Port int
	TLS  bool
}

// 定义配置项
// 1. 方法
type ServerOption func(*Server)

// 2. 配置项
func WithHost(host string) ServerOption {
	return func(s *Server) {
		s.Host = host
	}
}

func WithPort(port int) ServerOption {
	return func(s *Server) {
		s.Port = port
	}
}

func WithTLS(tls bool) ServerOption {
	return func(s *Server) {
		s.TLS = tls
	}
}

func NewServer(opts ...ServerOption) *Server {
	// server 初始化
	server := &Server{
		Host: "127.0.0.1",
		Port: 8080,
		TLS:  false,
	}
	for _, opt := range opts {
		opt(server)
	}
	return server
}

func TestNewServer(t *testing.T) {
	server := NewServer()
	fmt.Println(server.Host)
	fmt.Println(server.Port)
	fmt.Println(server.TLS)

	server = NewServer(WithHost("localhost"), WithPort(8081), WithTLS(true))
	fmt.Println(server.Host)
	fmt.Println(server.Port)
	fmt.Println(server.TLS)

	// assert.Equal(t, server.Host, "127.0.0.1")
}

package server

import (
	"fmt"
	"testing"

	"github.com/valyala/fasthttp"
)

// fasthttp
func TestHTTPServer(t *testing.T) {
	// 通过fasthttp请求
	status, body, err := fasthttp.Get(nil, "http://localhost:8000/api/articles")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("status: ", status)
	fmt.Println("body: ", string(body))
}

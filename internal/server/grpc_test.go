package server

import (
	"context"
	"fmt"
	"testing"

	v1 "kratos-realworld/api/realworld/v1"

	"google.golang.org/grpc"
)

func TestGRPCServer(t *testing.T) {
	conn, err := grpc.NewClient("localhost:9000", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := v1.NewRealWorldClient(conn)
	resp, err := client.ListArticles(context.Background(), &v1.ListArticlesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("resp: ", resp)
}

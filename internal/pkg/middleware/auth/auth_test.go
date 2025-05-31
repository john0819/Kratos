package auth

import (
	"fmt"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token := GenerateToken("hello", 123456)
	fmt.Printf("token: %v\n", token)
}

package auth

import (
	"fmt"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token := GenerateToken("hello", "2233")
	fmt.Printf("token: %v\n", token)
}

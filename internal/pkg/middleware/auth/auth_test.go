package auth

import (
	"fmt"
	"testing"
)

const secret = "Kn1GEInldSSoQJc/x7F/000D++yWRPvz7Bnq2K+m5T0="

func TestGenerateToken(t *testing.T) {
	token := GenerateToken(secret, 233)
	fmt.Printf("token: %v\n", token)
}

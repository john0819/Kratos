package auth

import (
	"fmt"
	"testing"
)

// john的token = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjE0NDQ0Nzg0MDAsInVzZXJpZCI6MX0.kWZBczk8md9LTF9WXBGs7sqSNgdwzgoQxeGVm66-eNs
// wong的token = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjE0NDQ0Nzg0MDAsInVzZXJpZCI6M30.DfagSolhdeMu07G5Gbw6Zf4XAA7B-_YtPuy7Nu02qD0

const secret = "Kn1GEInldSSoQJc/x7F/000D++yWRPvz7Bnq2K+m5T0="

func TestGenerateToken(t *testing.T) {
	token := GenerateToken(secret, 233)
	fmt.Printf("token: %v\n", token)
}

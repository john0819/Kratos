package biz

import (
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestHashPassword(t *testing.T) {
	hash := hashPassword("123456")
	fmt.Printf("hash: %v\n", string(hash))
}

func TestVerifyPassword(t *testing.T) {
	assert.Equal(t, true, verifyPassword("123456", "$2a$10$yHkYqPfmpCnIs8wEKN./u./CQ.pHxux6fa06VwnkvdZiRWiFjMsCS"))
	assert.NotEqual(t, true, verifyPassword("123", "$2a$10$yHkYqPfmpCnIs8wEKN./u./CQ.pHxux6fa06VwnkvdZiRWiFjMsCS"))
	assert.NotEqual(t, true, verifyPassword("123456", "/u./CQ.pHxux6fa06VwnkvdZiRWiFjMsCS"))
}

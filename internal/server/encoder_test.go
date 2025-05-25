package server

import (
	"encoding/json"
	"fmt"
	"kratos-realworld/internal/errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError(t *testing.T) {
	a := &errors.HTTPError{
		Errors: map[string][]string{
			"name": {"name is required"},
		},
	}
	b, err := json.Marshal(a)
	assert.NoError(t, err)
	fmt.Println(string(b))
}

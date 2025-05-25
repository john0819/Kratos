package errors

import (
	"encoding/json"
	"errors"
)

func NewHTTPError(code int, filed string, detail string) *HTTPError {
	return &HTTPError{
		Errors: map[string][]string{
			filed: {detail},
		},
		Code: code,
	}
}

// 标准化 错误结构体
type HTTPError struct {
	Errors map[string][]string `json:"errors"`
	Code   int                 `json:"-"`
}

func (e *HTTPError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// 把error类型转换成自定义的错误类型HTTPError
func FromError(err error) *HTTPError {
	if err == nil {
		return nil
	}

	var se *HTTPError
	if errors.As(err, &se) {
		return se
	}

	return &HTTPError{}
}

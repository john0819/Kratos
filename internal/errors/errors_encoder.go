package errors

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/errors"
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

	// 自定义的错误类型进行转换
	var se *HTTPError
	if errors.As(err, &se) {
		return se
	}

	// kratos的错误进行转换
	// transport层做grpc和http的错误转换
	if se := new(errors.Error); errors.As(err, &se) {
		return NewHTTPError(int(se.Code), se.Reason, se.Message)
	}

	return &HTTPError{}
}

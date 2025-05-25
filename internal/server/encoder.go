package server

import (
	"kratos-realworld/internal/errors"
	nethttp "net/http"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// 错误编码器 - 写数据到输出
// 统一的错误处理机制 - 处理框架返回的错误error
func errorEncoder(w nethttp.ResponseWriter, r *nethttp.Request, err error) {
	se := errors.FromError(err)
	codec, _ := http.CodecForRequest(r, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		w.WriteHeader(nethttp.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/"+codec.Name())
	w.WriteHeader(se.Code)
	_, _ = w.Write(body)
}

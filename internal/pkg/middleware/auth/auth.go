package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/golang-jwt/jwt/v4"
)

const (
	tokenWord = "Token"
)

func GenerateToken(secret, username string) string {
	// claim - payload部分
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"nbf":      time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		panic(err)
	}

	return tokenString
}

// server层的middleware 做auth鉴权 - path前缀来判断是否需要鉴权
// path前缀 - 白名单 - server层做的
func JWTAuth(secret string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				tokenString := tr.RequestHeader().Get("Authorization")

				if tokenString == "" {
					return nil, errors.New("token is required")
				}

				// 切割字符串 token的格式组成
				auths := strings.SplitN(tokenString, " ", 2)
				if len(auths) != 2 || !strings.EqualFold(auths[0], tokenWord) {
					return nil, errors.New("invalid token")
				}
				tokenString = auths[1]

				// 解析token后做鉴权校验
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte(secret), nil
				}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
				if err != nil {
					return nil, err
				}

				// jwt中的payload部分 - 数据
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					fmt.Printf("claims: %v\n", claims["username"])
				} else {
					return nil, errors.New("invalid token")
				}
			}
			return handler(ctx, req)
		}
	}
}

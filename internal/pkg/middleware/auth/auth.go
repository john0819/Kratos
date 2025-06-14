package auth

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/golang-jwt/jwt/v4"
)

const (
	tokenWord = "Token"
)

// 可选鉴权接口
var optionalAuthRouters = map[string]struct{}{
	"/realworld.v1.RealWorld/GetProfile":   {},
	"/realworld.v1.RealWorld/GetComments":  {},
	"/realworld.v1.RealWorld/ListArticles": {},
}

// 在context里面存储用户信息-uid
// learn: 专门用一个字段来存储用户信息
// 使用空结构体 - 唯一的key并且不占内存
var currentUserKey struct{}

type CurrentUser struct {
	UserID uint
}

// 生成token, 用户的username放在jwt中
func GenerateToken(secret string, userid uint) string {
	// claim - payload部分
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": userid,
		"nbf":    time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		panic(err)
	}

	return tokenString
}

// server层的middleware 做auth鉴权 - path前缀来判断是否需要鉴权
// path前缀 - 白名单 - server层做的
// 可选鉴权接口 需要在这层做判断 - 有token则判断， 没有则跳过
func JWTAuth(secret string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				// 获取当前接口路径
				currentPath := tr.Operation()
				// 获取token
				tokenString := tr.RequestHeader().Get("Authorization")

				if tokenString == "" {
					// 可选鉴权接口 - 没有token则跳过
					if _, ok := optionalAuthRouters[currentPath]; ok {
						return handler(ctx, req)
					}
					return nil, errors.Unauthorized("UNAUTHORIZED", "token is required")
				}

				// 切割字符串 token的格式组成
				auths := strings.SplitN(tokenString, " ", 2)
				if len(auths) != 2 || !strings.EqualFold(auths[0], tokenWord) {
					return nil, errors.Unauthorized("UNAUTHORIZED", "invalid token")
				}
				tokenString = auths[1]

				// 解析token后做鉴权校验
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte(secret), nil
				}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
				if err != nil {
					return nil, errors.Unauthorized("UNAUTHORIZED", "invalid token")
				}

				// jwt中的payload部分 - 数据
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					// fmt.Printf("claims: %v\n", claims["userid"])
					if u, ok := claims["userid"]; ok {
						// 鉴权通过后, 把user信息塞入ctx中 - 方便后续获取鉴权用户信息uid唯一性
						// jwt 解析时会解析为float64 - 断言为float64以后进行转换
						ctx = WithContext(ctx, &CurrentUser{UserID: uint(u.(float64))})
					}
				} else {
					return nil, errors.Unauthorized("UNAUTHORIZED", "invalid token claims")
				}
			}
			return handler(ctx, req)
		}
	}
}

// 获取ctx中的user信息
func FromContext(ctx context.Context) (*CurrentUser, bool) {
	u, ok := ctx.Value(currentUserKey).(*CurrentUser)
	return u, ok
}

// 设置ctx中的user信息
func WithContext(ctx context.Context, user *CurrentUser) context.Context {
	return context.WithValue(ctx, currentUserKey, user)
}

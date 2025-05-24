package service

import (
	v1 "kratos-realworld/api/realworld/v1"
	"kratos-realworld/internal/biz"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewRealWorldService)

type RealWorldService struct {
	// [1] grpc的服务 实现了SayHello方法, 从而可以grpc调用
	v1.UnimplementedRealWorldServer

	ur *biz.UserUsecase
	uc *biz.SocialUsecase
}

// [1] 通过new的方式进行实例初始化
func NewRealWorldService(ur *biz.UserUsecase, uc *biz.SocialUsecase) *RealWorldService {
	return &RealWorldService{uc: uc, ur: ur}
}

// [1] service层实现所有api的方法

package service

import (
	"context"

	v1 "kratos-realworld/api/realworld/v1"
)

// service层 - handler对api接口进行实现 具体的业务逻辑在业务层/biz 数据操作在数据层/data
func (s *RealWorldService) Login(ctx context.Context, in *v1.LoginRequest) (*v1.UserResponse, error) {
	return &v1.UserResponse{
		User: &v1.UserResponse_User{
			Username: "john",
		},
	}, nil
}

func (s *RealWorldService) Register(ctx context.Context, in *v1.RegisterRequest) (*v1.UserResponse, error) {
	return &v1.UserResponse{
		User: &v1.UserResponse_User{
			Username: "john",
		},
	}, nil
}

func (s *RealWorldService) GetCurrentUser(ctx context.Context, in *v1.GetCurrentUserRequest) (*v1.UserResponse, error) {
	return &v1.UserResponse{}, nil
}

func (s *RealWorldService) UpdateUser(ctx context.Context, in *v1.UpdateUserRequest) (*v1.UserResponse, error) {
	return &v1.UserResponse{}, nil
}

func (s *RealWorldService) GetProfile(ctx context.Context, in *v1.GetProfileRequest) (*v1.ProfileResponse, error) {
	return &v1.ProfileResponse{}, nil
}

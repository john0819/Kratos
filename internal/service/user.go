package service

import (
	"context"

	v1 "kratos-realworld/api/realworld/v1"
	"kratos-realworld/internal/errors"
)

// service层 - handler对api接口进行实现 具体的业务逻辑在业务层/biz 数据操作在数据层/data
func (s *RealWorldService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.UserResponse, error) {
	if len(req.User.Email) == 0 {
		// 函数返回的类型的是 error, 需要通过错误转换器来转换
		return nil, errors.NewHTTPError(422, "email", "can not be empty")
	}
	if len(req.User.Password) == 0 {
		return nil, errors.NewHTTPError(422, "password", "can not be empty")
	}

	return &v1.UserResponse{
		User: &v1.UserResponse_User{
			Username: "john",
		},
	}, nil
}

// service层要调用biz层
func (s *RealWorldService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.UserResponse, error) {
	user, err := s.ur.Register(ctx, req.User.Username, req.User.Email, req.User.Password)
	if err != nil {
		return nil, err
	}
	return &v1.UserResponse{
		User: &v1.UserResponse_User{
			Username: user.Username,
			Email:    user.Email,
			Token:    user.Token,
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

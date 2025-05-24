package biz

import (
	"context"

	v1 "kratos-realworld/api/realworld/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	// ErrUserNotFound is user not found.
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

type User struct {
	ID       int64
	Username string
	Email    string
	Password string
}

// user - profile / follow / unfollow
type UserRepo interface {
	CreateUser(ctx context.Context, user *User) error
}

type ProfileRepo interface {
}

// GreeterUsecase is a Greeter usecase.
type UserUsecase struct {
	ur  UserRepo
	pr  ProfileRepo
	log *log.Helper
}

func NewUserUsecase(ur UserRepo,
	pr ProfileRepo,
	logger log.Logger,
) *UserUsecase {
	return &UserUsecase{ur: ur, pr: pr, log: log.NewHelper(logger)}
}

func (uc *UserUsecase) Register(ctx context.Context, user *User) error {
	if err := uc.ur.CreateUser(ctx, user); err != nil {
		return err
	}
	return nil
}

func (uc *UserUsecase) Login(ctx context.Context) error {
	return nil
}

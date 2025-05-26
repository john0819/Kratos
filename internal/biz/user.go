package biz

import (
	"context"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
)

// 请求
type User struct {
	Email        string
	Username     string
	Bio          string
	Image        string
	PasswordHash string
}

// 响应
type UserLogin struct {
	Email    string
	Username string
	Token    string
	Bio      string
	Image    string
}

// hash password - 数据库中存储hash加密的pwd
func hashPassword(pwd string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

// verify password - 登录时验证密码
func verifyPassword(pwd string, hash string) bool {
	// 第一个明文密码, 第二个为数据库存的hash密码
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
	return err == nil
}

// user - profile / follow / unfollow
type UserRepo interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}

type ProfileRepo interface {
}

// GreeterUsecase is a Greeter usecase.
type UserUsecase struct {
	ur   UserRepo
	pr   ProfileRepo
	log  *log.Helper
	jwtc *conf.JWT
}

func NewUserUsecase(ur UserRepo,
	pr ProfileRepo,
	logger log.Logger,
	jwtc *conf.JWT,
) *UserUsecase {
	return &UserUsecase{ur: ur, pr: pr, log: log.NewHelper(logger), jwtc: jwtc}
}

func (uc *UserUsecase) generateToken(username string) string {
	return auth.GenerateToken(uc.jwtc.Secret, username)
}

func (uc *UserUsecase) Register(ctx context.Context, username string, email string, password string) (*UserLogin, error) {
	u := &User{
		Username:     username,
		Email:        email,
		PasswordHash: hashPassword(password),
	}

	if err := uc.ur.CreateUser(ctx, u); err != nil {
		return nil, err
	}

	// 通过jwt生成token并返回
	return &UserLogin{
		Email:    email,
		Username: username,
		Token:    uc.generateToken(username),
	}, nil
}

func (uc *UserUsecase) Login(ctx context.Context, email string, password string) (*UserLogin, error) {
	// invalid 逻辑放在biz层
	if len(email) == 0 {
		return nil, errors.New(422, "email", "can not be empty")
	}
	if len(password) == 0 {
		return nil, errors.New(422, "password", "can not be empty")
	}

	u, err := uc.ur.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if !verifyPassword(password, u.PasswordHash) {
		return nil, errors.Unauthorized("password", "invalid password")
	}
	return &UserLogin{
		Email:    u.Email,
		Username: u.Username,
		Token:    uc.generateToken(u.Username),
		Bio:      u.Bio,
		Image:    u.Image,
	}, nil
}

package biz

import (
	"context"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
)

// biz层定义数据结构 - 用户data层和service层
// 请求 - data层的入参
type User struct {
	ID           uint
	Email        string
	Username     string
	Bio          string
	Image        string
	PasswordHash string
}

// 更新用户数据
type UserUpdate struct {
	Email    string
	Password string
	Username string
	Bio      string
	Image    string
}

// 响应 - data层的响应
type UserLogin struct {
	Email    string
	Username string
	Token    string
	Bio      string
	Image    string
}

type ProfileResp struct {
	Username  string
	Bio       string
	Image     string
	Following bool
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
	GetUserByID(ctx context.Context, uid uint) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
}

type ProfileRepo interface {
	GetProfileByUsername(ctx context.Context, username string) (*ProfileResp, error)
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

func (uc *UserUsecase) generateToken(uid uint) string {
	return auth.GenerateToken(uc.jwtc.Secret, uid)
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
		Token:    uc.generateToken(u.ID),
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
	// 比对登录密码 和 数据库对应的hash密码
	if !verifyPassword(password, u.PasswordHash) {
		return nil, errors.Unauthorized("password", "invalid password")
	}
	return &UserLogin{
		Email:    u.Email,
		Username: u.Username,
		Token:    uc.generateToken(u.ID),
		Bio:      u.Bio,
		Image:    u.Image,
	}, nil
}

func (uc *UserUsecase) GetCurrentUser(ctx context.Context) (*User, error) {
	// 从ctx中获取用户的uid
	uidCtx, _ := auth.FromContext(ctx)
	user, err := uc.ur.GetUserByID(ctx, uidCtx.UserID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// 优化点: 对所有字段都可以进行更新, 通过uid进行判断, 不需要email做限制
func (uc *UserUsecase) UpdateUserInfo(ctx context.Context, userUpdate *UserUpdate) (*UserLogin, error) {
	// 1. 先获取数据库中的内容
	uidCtx, _ := auth.FromContext(ctx)
	userFromDB, err := uc.ur.GetUserByID(ctx, uidCtx.UserID)
	if err != nil {
		return nil, err
	}
	// 2. 通过数据库中的内容修改, 再去update数据库
	if userUpdate.Email != "" {
		userFromDB.Email = userUpdate.Email
	}
	if userUpdate.Password != "" {
		userFromDB.PasswordHash = hashPassword(userUpdate.Password)
	}
	if userUpdate.Username != "" {
		userFromDB.Username = userUpdate.Username
	}
	if userUpdate.Bio != "" {
		userFromDB.Bio = userUpdate.Bio
	}
	if userUpdate.Image != "" {
		userFromDB.Image = userUpdate.Image
	}
	// 3. 更新数据库
	userFromDB, err = uc.ur.UpdateUser(ctx, userFromDB)
	if err != nil {
		return nil, err
	}
	return &UserLogin{
		Email:    userFromDB.Email,
		Username: userFromDB.Username,
		Token:    uc.generateToken(userFromDB.ID),
		Bio:      userFromDB.Bio,
		Image:    userFromDB.Image,
	}, nil
}

func (uc *UserUsecase) GetProfile(ctx context.Context, username string) (*ProfileResp, error) {
	profile, err := uc.pr.GetProfileByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

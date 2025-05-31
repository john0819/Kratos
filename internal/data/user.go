package data

import (
	"context"
	"strings"

	"kratos-realworld/internal/biz"
	e "kratos-realworld/internal/errors"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// data层定义数据库中的数据结构
type User struct {
	gorm.Model
	Email        string `gorm:"size:500;unique"`
	Username     string `gorm:"size:500"`
	Bio          string `gorm:"size:1000"`
	Image        string `gorm:"size:1000"`
	PasswordHash string `gorm:"size:500"`
	Following    uint32
}

// 具体实现 biz层的interface
type userRepo struct {
	data *Data
	log  *log.Helper
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userRepo) CreateUser(ctx context.Context, user *biz.User) error {
	u := User{
		Email:        user.Email,
		Username:     user.Username,
		Bio:          user.Bio,
		Image:        user.Image,
		PasswordHash: user.PasswordHash,
	}
	if err := r.data.db.Create(&u).Error; err != nil {
		// 检查错误是否为重复的key
		if strings.Contains(err.Error(), "Duplicate entry") {
			return e.NewHTTPError(400, "email", "email already exists")
		}
		return e.NewHTTPError(500, "database", "database error")
	}
	// uid需要返回到biz层 - token需要
	user.ID = u.ID
	return nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*biz.User, error) {
	u := new(User)
	result := r.data.db.Where("email = ?", email).First(u)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("user", "not found by email")
	}

	return &biz.User{
		ID:           u.ID,
		Email:        u.Email,
		Username:     u.Username,
		Bio:          u.Bio,
		Image:        u.Image,
		PasswordHash: u.PasswordHash,
	}, nil

}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*biz.User, error) {
	return nil, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, uid uint) (*biz.User, error) {
	u := new(User)
	result := r.data.db.Where("id = ?", uid).First(u)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("user", "not found by id")
	}
	return &biz.User{
		// uid返回是为了做修改的时候 能够确保知道是哪个uid
		ID:       u.ID,
		Email:    u.Email,
		Username: u.Username,
		Bio:      u.Bio,
		Image:    u.Image,
	}, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, user *biz.User) (*biz.User, error) {
	// uid是唯一的
	u := new(User)
	// 1. 先找到要修改的用户
	if err := r.data.db.Where("id = ?", user.ID).First(u).Error; err != nil {
		return nil, err
	}
	// 2. 更新用户信息
	err := r.data.db.Model(&u).Updates(User{
		Email:        user.Email,
		Username:     user.Username,
		Bio:          user.Bio,
		Image:        user.Image,
		PasswordHash: user.PasswordHash,
	}).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, errors.BadRequest("email", "email already exists")
		}
		return nil, err
	}

	// 返回更新后内容
	return &biz.User{
		ID:           u.ID,
		Email:        u.Email,
		Username:     u.Username,
		Bio:          u.Bio,
		Image:        u.Image,
		PasswordHash: u.PasswordHash,
	}, nil
}

type profileRepo struct {
	data *Data
	log  *log.Helper
}

func NewProfileRepo(data *Data, logger log.Logger) biz.ProfileRepo {
	return &profileRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

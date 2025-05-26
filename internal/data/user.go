package data

import (
	"context"
	"strings"

	"kratos-realworld/internal/biz"
	e "kratos-realworld/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email        string `gorm:"size:500;unique"`
	Username     string `gorm:"size:500"`
	Bio          string `gorm:"size:1000"`
	Image        string `gorm:"size:1000"`
	PasswordHash string `gorm:"size:500"`
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
	return nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*biz.User, error) {
	var u User
	if err := r.data.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, e.NewHTTPError(400, "user", "user not found")
	}
	return &biz.User{
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

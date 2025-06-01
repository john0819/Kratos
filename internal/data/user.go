package data

import (
	"context"
	"strings"

	"kratos-realworld/internal/biz"
	e "kratos-realworld/internal/errors"
	"kratos-realworld/internal/pkg/middleware/auth"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// data层定义数据库中的数据结构
// mark: 由于username需要用来做profile的几个接口方法参数, 所以要做成唯一的
type User struct {
	gorm.Model
	Email        string `gorm:"size:500;unique"`
	Username     string `gorm:"size:500;unique"`
	Bio          string `gorm:"size:1000"`
	Image        string `gorm:"size:1000"`
	PasswordHash string `gorm:"size:500"`
}

// follow表 - 关注id和被关注id
type Follow struct {
	gorm.Model
	FollowerID  uint `gorm:"index"` // 关注者的id - 粉丝
	FollowingID uint `gorm:"index"` // 被关注者的id - 博主
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
			if strings.Contains(err.Error(), "username") {
				return e.NewHTTPError(400, "username", "username already exists")
			}
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
			if strings.Contains(err.Error(), "username") {
				return nil, errors.BadRequest("username", "username already exists")
			}
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

func (p *profileRepo) GetProfileByUsername(ctx context.Context, username string) (*biz.ProfileResp, error) {
	u := new(User)
	// 1. 获取username对应的数据
	if err := p.data.db.Where("username = ?", username).First(u).Error; err != nil {
		return nil, err
	}
	// 2. 查看当前用户是否关注该博主username
	var following bool
	currentUser, ok := auth.FromContext(ctx)
	if ok {
		var count int64
		err := p.data.db.Model(&Follow{}).
			Where("follower_id = ? AND following_id = ?", currentUser.UserID, u.ID).
			Count(&count).Error
		if err != nil {
			return nil, err
		}
		following = count > 0
	}

	return &biz.ProfileResp{
		Username:  u.Username,
		Bio:       u.Bio,
		Image:     u.Image,
		Following: following,
	}, nil
}

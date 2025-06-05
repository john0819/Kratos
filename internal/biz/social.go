package biz

import (
	"context"
	"kratos-realworld/internal/pkg/middleware/auth"
	"regexp"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// slug 转换特殊字符为-
func slugify(title string) string {
	re, _ := regexp.Compile(`[^\w]`)
	return strings.ToLower(re.ReplaceAllString(title, "-"))
}

// 请求和响应结构体定义
type Article struct {
	ID             uint
	Slug           string // title一般不友好对于url的path来说, 通过slug以 - 来连接解决
	Title          string
	Description    string
	Body           string
	TagList        []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Favorited      bool
	FavoritesCount uint32

	// 作者的uid 从请求获取
	AuthorID uint
	// 作者的profile
	Author *ProfileResp
}

// social - article / comment / tag
type ArticleRepo interface {
	CreateArticle(ctx context.Context, article *Article) (*Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (*Article, error)
	DeleteArticleBySlug(ctx context.Context, slug string) error
	UpdateArticle(ctx context.Context, article *Article) (*Article, error)
}

type CommentRepo interface {
}

type TagRepo interface {
}

// GreeterUsecase is a Greeter usecase.
type SocialUsecase struct {
	ar  ArticleRepo
	cr  CommentRepo
	tr  TagRepo
	log *log.Helper
}

func NewSocialUsecase(ar ArticleRepo,
	cr CommentRepo,
	tr TagRepo,
	logger log.Logger,
) *SocialUsecase {
	return &SocialUsecase{ar: ar, cr: cr, tr: tr, log: log.NewHelper(logger)}
}

func (uc *SocialUsecase) CreateArticle(ctx context.Context, a *Article) (*Article, error) {
	// 请求已经包含title, des, body, tags
	// biz层获取uid, 转换title为slug
	currentUser, _ := auth.FromContext(ctx)
	currentUid := currentUser.UserID
	a.AuthorID = currentUid
	a.Slug = slugify(a.Title)

	// data层创建文章
	article, err := uc.ar.CreateArticle(ctx, a)
	if err != nil {
		return nil, err
	}

	// favorited 和 author是否follow这层传出去
	article.Favorited = false
	// 指针需要进行保护
	if article.Author != nil {
		article.Author.Following = false
	}

	return article, nil
}

func (uc *SocialUsecase) GetArticle(ctx context.Context, slug string) (*Article, error) {
	article, err := uc.ar.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (uc *SocialUsecase) DeleteArticle(ctx context.Context, slug string) error {
	return nil
}

func (uc *SocialUsecase) UpdateArticle(ctx context.Context, article *Article) (*Article, error) {
	return nil, nil
}

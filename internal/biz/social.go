package biz

import (
	"context"
	"kratos-realworld/internal/pkg/middleware/auth"
	"kratos-realworld/internal/pkg/utils"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// 权限判断
func verifyAuthor(ctx context.Context, article *Article, currentUid uint) bool {
	return currentUid == article.AuthorID
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
	a.Slug = utils.Slugify(a.Title)

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
	// 获取文章
	uc.log.Infof("delete article by slug: %s", slug)
	a, err := uc.ar.GetArticleBySlug(ctx, slug)
	if err != nil {
		uc.log.Errorf("get article by slug error: %v", err)
		return err
	}

	// 只有作者才有权限删除
	currentUser, _ := auth.FromContext(ctx)
	currentUid := currentUser.UserID
	if !verifyAuthor(ctx, a, currentUid) {
		return errors.Forbidden("FORBIDDEN", "you are not the author of this article")
	}

	// 删除文章
	return uc.ar.DeleteArticleBySlug(ctx, slug)
}

func (uc *SocialUsecase) UpdateArticle(ctx context.Context, article *Article) (*Article, error) {
	uc.log.Infof("update article by slug: %s", article.Slug)
	// 获取文章
	a, err := uc.ar.GetArticleBySlug(ctx, article.Slug)
	if err != nil {
		return nil, err
	}

	// 验证是否为作者
	currentUser, _ := auth.FromContext(ctx)
	currentUid := currentUser.UserID
	if !verifyAuthor(ctx, a, currentUid) {
		return nil, errors.Forbidden("FORBIDDEN", "you are not the author of this article")
	}

	// 需要更新的请求
	updateArticle := &Article{
		Slug:        article.Slug,
		Title:       article.Title,
		Description: article.Description,
		Body:        article.Body,
		TagList:     article.TagList,
	}

	article, err = uc.ar.UpdateArticle(ctx, updateArticle)
	if err != nil {
		uc.log.Errorf("update article error: %v", err)
		return nil, err
	}

	return article, nil
}

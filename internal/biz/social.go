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

// 关于当前用户与文章之间的关系 收藏关系
func (uc *SocialUsecase) getArticleFavoritedByUid(ctx context.Context, articles []*Article, currentUid uint) ([]*Article, error) {
	aids := make([]uint, 0)
	for _, article := range articles {
		aids = append(aids, article.ID)
	}
	favoriteMap, err := uc.ar.GetIsFavorited(ctx, aids, currentUid)
	if err != nil {
		return nil, err
	}
	for _, article := range articles {
		article.Favorited = favoriteMap[article.ID]
	}
	return articles, nil
}

// 关于当前用户与文章作者之间的 关注关系
func (uc *SocialUsecase) getArticleAuthorFollowedByUid(ctx context.Context, articles []*Article, currentUid uint) ([]*Article, error) {
	uc.log.Infof("关于当前用户与文章作者之间的 关注关系")
	authorIds := make([]uint, 0)
	for _, article := range articles {
		authorIds = append(authorIds, article.AuthorID)
	}

	// 批量查询所有作者的关注状态
	followingMap, err := uc.ar.GetOneIsFollowingAnother(ctx, currentUid, authorIds)
	if err != nil {
		return nil, err
	}

	// 将关注状态映射到每篇文章
	for _, article := range articles {
		article.Author.Following = followingMap[article.AuthorID]
	}

	return articles, nil
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

type Comment struct {
	ID        uint
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Author    *ProfileResp

	// 传给data层的时候, 需要将authorID和articleID传给data层 - 方便查询
	// author
	AuthorID uint
	// article
	ArticleID uint
}

type Tag string

// social - article / comment / tag
type ArticleRepo interface {
	CreateArticle(ctx context.Context, article *Article) (*Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (*Article, error)
	DeleteArticleBySlug(ctx context.Context, slug string) error
	UpdateArticle(ctx context.Context, article *Article) (*Article, error)
	GetArticleByAid(ctx context.Context, aid uint) (*Article, error)

	FavoriteArticle(ctx context.Context, aid uint, uid uint) error
	UnfavoriteArticle(ctx context.Context, aid uint, uid uint) error
	GetIsFavorited(ctx context.Context, aids []uint, uid uint) (map[uint]bool, error)

	ListArticlesByOptions(ctx context.Context, options *ListOptions) ([]*Article, error)
	GetOneIsFollowingAnother(ctx context.Context, uid_1 uint, uids []uint) (map[uint]bool, error)
}

type CommentRepo interface {
	AddComment(ctx context.Context, c *Comment) (*Comment, error)
	DeleteCommentByID(ctx context.Context, id uint) error
	GetCommentsByID(ctx context.Context, cid uint) ([]*Comment, error)
}

type TagRepo interface {
	GetTags(ctx context.Context) ([]Tag, error)
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
	uc.log.Infof("get article by slug: %s", slug)
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

	// 获取是否收藏
	favoriteMap, err := uc.ar.GetIsFavorited(ctx, []uint{article.ID}, currentUid)
	if err != nil {
		return nil, err
	}
	article.Favorited = favoriteMap[article.ID]

	return article, nil
}

func (uc *SocialUsecase) FavoriteArticle(ctx context.Context, slug string) (*Article, error) {
	uc.log.Infof("favorite article by slug: %s", slug)
	// 获取文章
	a, err := uc.ar.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// 获取用户
	currentUser, _ := auth.FromContext(ctx)
	currentUid := currentUser.UserID

	// 添加喜欢
	err = uc.ar.FavoriteArticle(ctx, a.ID, currentUid)
	if err != nil {
		return nil, err
	}

	// 更新文章信息
	a, err = uc.ar.GetArticleByAid(ctx, a.ID)
	if err != nil {
		return nil, err
	}

	// 获取是否收藏
	favoriteMap, err := uc.ar.GetIsFavorited(ctx, []uint{a.ID}, currentUid)
	if err != nil {
		return nil, err
	}
	a.Favorited = favoriteMap[a.ID]

	return a, nil
}

func (uc *SocialUsecase) UnfavoriteArticle(ctx context.Context, slug string) (*Article, error) {
	uc.log.Infof("unfavorite article by slug: %s", slug)
	// 获取文章
	a, err := uc.ar.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// 获取用户
	currentUser, _ := auth.FromContext(ctx)
	currentUid := currentUser.UserID

	err = uc.ar.UnfavoriteArticle(ctx, a.ID, currentUid)
	if err != nil {
		return nil, err
	}

	// 更新文章信息
	a, err = uc.ar.GetArticleByAid(ctx, a.ID)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// 查询文章
func (uc *SocialUsecase) ListArticles(ctx context.Context, opts ...ListOption) ([]*Article, error) {
	uc.log.Infof("list articles by opts: %v", opts)
	// 查询参数 - 根据service层进行配置
	options := NewListOptions(opts...)
	// 当前用户 - 得看是否登录
	currentUser, _ := auth.FromContext(ctx)
	if currentUser != nil {
		currentUid := currentUser.UserID
		options.CurrentUid = currentUid
	}
	articles, err := uc.ar.ListArticlesByOptions(ctx, options)
	if err != nil {
		return nil, err
	}

	// 如果有鉴权登录, 查询aid和uid的收藏关系 + uid和authorId的follow关系
	if currentUser != nil {
		currentUid := currentUser.UserID
		articles, err = uc.getArticleFavoritedByUid(ctx, articles, currentUid)
		if err != nil {
			return nil, err
		}
		articles, err = uc.getArticleAuthorFollowedByUid(ctx, articles, currentUid)
		if err != nil {
			return nil, err
		}
	}

	return articles, nil
}

// 查询文章 - 登录用户与其关注用户的关系
func (uc *SocialUsecase) FeedArticles(ctx context.Context, opts ...ListOption) ([]*Article, error) {
	uc.log.Info("feed artile by opts: %v", opts)
	options := NewListOptions(opts...)
	currentUser, _ := auth.FromContext(ctx)
	var currentUid uint
	if currentUser != nil {
		currentUid = currentUser.UserID
		options.CurrentUid = currentUid
	}
	uc.log.Infof("feed articles by uid: %v", options.CurrentUid)
	articles, err := uc.ar.ListArticlesByOptions(ctx, options)
	if err != nil {
		return nil, err
	}

	// uid和aid的收藏关系
	articles, err = uc.getArticleFavoritedByUid(ctx, articles, currentUid)
	if err != nil {
		return nil, err
	}

	// uid和authorId的follow关系
	articles, err = uc.getArticleAuthorFollowedByUid(ctx, articles, currentUid)
	if err != nil {
		return nil, err
	}

	return articles, nil
}

func (uc *SocialUsecase) AddComment(ctx context.Context, slug string, c *Comment) (*Comment, error) {
	uc.log.Infof("add comment by slug: %s", slug)

	// 评论用户的id
	currentUser, _ := auth.FromContext(ctx)
	currentUid := currentUser.UserID

	// 评论的文章id
	a, err := uc.ar.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	c.ArticleID = a.ID
	c.AuthorID = currentUid

	comment, err := uc.cr.AddComment(ctx, c)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (uc *SocialUsecase) DeleteComment(ctx context.Context, slug string, id uint) error {
	uc.log.Infof("delete comment by slug: %s, id: %d", slug, id)

	// 只有作者才能删除
	currentUser, _ := auth.FromContext(ctx)
	currentUid := currentUser.UserID
	a, err := uc.ar.GetArticleBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if !verifyAuthor(ctx, a, currentUid) {
		return errors.Forbidden("FORBIDDEN", "you are not the author of this article")
	}

	err = uc.cr.DeleteCommentByID(ctx, id)
	return err
}

func (uc *SocialUsecase) GetComments(ctx context.Context, slug string) ([]*Comment, error) {
	uc.log.Infof("get comments by slug: %s", slug)
	a, err := uc.ar.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	comments, err := uc.cr.GetCommentsByID(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (uc *SocialUsecase) GetTags(ctx context.Context) ([]Tag, error) {
	return uc.tr.GetTags(ctx)
}

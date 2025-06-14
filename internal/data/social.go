package data

import (
	"context"
	"kratos-realworld/internal/biz"
	"kratos-realworld/internal/pkg/utils"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 转换data.Article为biz.Article
func convertArticle(a Article) *biz.Article {
	return &biz.Article{
		ID:             a.ID,
		Slug:           a.Slug,
		Title:          a.Title,
		Description:    a.Description,
		Body:           a.Body,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
		FavoritesCount: a.FavoritesCount,
		TagList: func() []string {
			tags := make([]string, len(a.Tags))
			for i, tag := range a.Tags {
				tags[i] = tag.Name
			}
			return tags
		}(),
		AuthorID: a.AuthorID,
		Author: &biz.ProfileResp{
			ID:       a.Author.ID,
			Username: a.Author.Username,
			Bio:      a.Author.Bio,
			Image:    a.Author.Image,
		},
	}
}

// 收藏数量
func countFavorites(db *gorm.DB, aid uint) (uint32, error) {
	var count int64
	e := db.Model(&ArticleFavorite{}).Where("article_id = ?", aid).Count(&count).Error
	if e != nil {
		return 0, e
	}
	return uint32(count), nil
}

// 定义数据库表结构

// 文章表
type Article struct {
	gorm.Model
	Slug           string `gorm:"size:500;unique"`
	Title          string `gorm:"size:500"`
	Description    string `gorm:"size:1000"`
	Body           string `gorm:"size:10000"`
	Tags           []Tag  `gorm:"many2many:article_tags;constraint:OnDelete:CASCADE;"`
	AuthorID       uint
	Author         User // 关联user表
	FavoritesCount uint32
	Favorites      []ArticleFavorite `gorm:"constraint:OnDelete:CASCADE;"`
	Comments       []Comment         `gorm:"constraint:OnDelete:CASCADE;"`
}

// comment表
type Comment struct {
	gorm.Model
	ArticleID uint    // 关联article表
	Article   Article `gorm:"constraint:OnDelete:CASCADE;"`
	Body      string
	AuthorID  uint // 关联user表
	Author    User
}

// tag表
// 一个tag可以对应多个文章
type Tag struct {
	gorm.Model
	Name     string    `gorm:"size:500;uniqueIndex"`
	Articles []Article `gorm:"many2many:article_tags;"`
}

// 文章和用户的关联表
type ArticleFavorite struct {
	gorm.Model
	UserID    uint `gorm:"index:idx_user_article,unique"`
	ArticleID uint `gorm:"index:idx_user_article,unique;constraint:OnDelete:CASCADE;"`
}

// 文章和tag的关联表

type articleRepo struct {
	data *Data
	log  *log.Helper
}

func NewArticleRepo(data *Data, logger log.Logger) biz.ArticleRepo {
	return &articleRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (ar *articleRepo) CreateArticle(ctx context.Context, article *biz.Article) (*biz.Article, error) {
	// 组装tags, 并写入tag表
	tags := make([]Tag, len(article.TagList))
	for i, tagName := range article.TagList {
		tags[i] = Tag{
			Name: tagName,
		}
	}
	if len(tags) > 0 {
		err := ar.data.db.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&tags).Error
		if err != nil {
			return nil, err
		}
	}

	// 以防tag为0的情况（如果tag已经存在, 则不会创建, 所以需要查询） - fix
	var dbTags []Tag
	ar.data.db.Where("name IN ?", article.TagList).Find(&dbTags)

	a := Article{
		Slug:        article.Slug,
		Title:       article.Title,
		Description: article.Description,
		Body:        article.Body,
		Tags:        dbTags,
		AuthorID:    article.AuthorID,
		Author:      User{Model: gorm.Model{ID: article.AuthorID}},
	}

	err := ar.data.db.Create(&a).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			if strings.Contains(err.Error(), "slug") {
				return nil, errors.BadRequest("slug", "slug already exists")
			}
		}
		return nil, err
	}

	return convertArticle(a), nil
}

func (ar *articleRepo) GetArticleBySlug(ctx context.Context, slug string) (*biz.Article, error) {
	a := Article{}
	result := ar.data.db.Where("slug = ?", slug).Preload("Author").Preload("Tags").Preload("Favorites").First(&a)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("ARTICLE_NOT_FOUND", "article not found")
		}
		return nil, result.Error
	}

	article := convertArticle(a)
	// favorited count
	fc, err := countFavorites(ar.data.db, a.ID)
	if err != nil {
		return nil, err
	}
	article.FavoritesCount = fc
	return article, nil
}

func (ar *articleRepo) GetArticleByAid(ctx context.Context, aid uint) (*biz.Article, error) {
	a := Article{}
	err := ar.data.db.Model(&Article{}).Where("id = ?", aid).Preload("Author").Preload("Tags").Preload("Favorites").First(&a).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("ARTICLE_NOT_FOUND", "article not found")
		}
		return nil, err
	}
	// 收藏数量
	fc, err := countFavorites(ar.data.db, a.ID)
	if err != nil {
		return nil, err
	}
	a.FavoritesCount = fc

	return convertArticle(a), nil
}

func (ar *articleRepo) DeleteArticleBySlug(ctx context.Context, slug string) error {
	// 关联tag表和favorites表, 可以跟随删除
	return ar.data.db.Delete(&Article{}, "slug = ?", slug).Error
}

func (ar *articleRepo) UpdateArticle(ctx context.Context, article *biz.Article) (*biz.Article, error) {
	var dbArticle Article
	// 查到数据库中的文章内容
	err := ar.data.db.Model(&Article{}).Where("slug = ?", article.Slug).First(&dbArticle).Error
	if err != nil {
		return nil, err
	}

	// 更新文章内容
	if article.Title != "" {
		dbArticle.Title = article.Title
		dbArticle.Slug = utils.Slugify(article.Title)
	}
	if article.Description != "" {
		dbArticle.Description = article.Description
	}
	if article.Body != "" {
		dbArticle.Body = article.Body
	}
	if err := ar.data.db.Save(&dbArticle).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "slug") {
			return nil, errors.BadRequest("title", "title already exists")
		}
		return nil, err
	}

	// 更新tag - 需要更新关联
	if len(article.TagList) > 0 {
		// 先确保所有 tag 都存在
		tags := make([]Tag, len(article.TagList))
		for i, tagName := range article.TagList {
			tags[i] = Tag{Name: tagName}
		}
		ar.data.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tags)
		// 查出所有 tag 的完整记录
		var dbTags []Tag
		ar.data.db.Where("name IN ?", article.TagList).Find(&dbTags)
		// 更新关联
		if err := ar.data.db.Model(&dbArticle).Association("Tags").Replace(dbTags); err != nil {
			return nil, err
		}
	}

	return convertArticle(dbArticle), nil
}

// mark 这个接口既能收藏又能取消收藏
func (ar *articleRepo) FavoriteArticle(ctx context.Context, aid uint, uid uint) error {
	af := ArticleFavorite{
		UserID:    uid,
		ArticleID: aid,
	}

	// 获取文章
	a := Article{}
	err := ar.data.db.Model(&Article{}).Where("id = ?", aid).First(&a).Error
	if err != nil {
		return err
	}

	// 添加喜欢
	// 先查询
	result := ar.data.db.Where(&af).First(&ArticleFavorite{})
	// 没收藏过则收藏
	if result.RowsAffected == 0 {
		ar.log.Infof("favorite article by article_id: %d", aid)
		err := ar.data.db.Create(&af).Error
		if err != nil {
			return err
		}
	} else {
		// 收藏过则取消收藏
		// 采用物理删除
		ar.log.Infof("unfavorite article by article_id: %d", aid)
		err := ar.data.db.Unscoped().Where(&af).Delete(&ArticleFavorite{}).Error
		if err != nil {
			return err
		}
	}

	// todo: article里的收藏数量可以删除这个字段, 直接通过article_favorites表来计算
	fc, err := countFavorites(ar.data.db, aid)
	if err != nil {
		return err
	}
	a.FavoritesCount = fc
	ar.log.Infof("收藏数量: %d", a.FavoritesCount)

	// 更新文章
	err = ar.data.db.Model(&Article{}).Where("id = ?", aid).UpdateColumn("favorites_count", a.FavoritesCount).Error
	return err
}

// 只做取消收藏
func (ar *articleRepo) UnfavoriteArticle(ctx context.Context, aid uint, uid uint) error {
	af := ArticleFavorite{
		UserID:    uid,
		ArticleID: aid,
	}

	a := Article{}
	err := ar.data.db.Model(&Article{}).Where("id = ?", aid).First(&a).Error
	if err != nil {
		return nil
	}

	result := ar.data.db.Unscoped().Where(&af).Delete(&ArticleFavorite{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("FAVORITE_NOT_FOUND", "you do notfavorite this article")
	}

	fc, err := countFavorites(ar.data.db, aid)
	if err != nil {
		return err
	}
	a.FavoritesCount = fc
	err = ar.data.db.Model(&Article{}).Where("id = ?", aid).UpdateColumn("favorites_count", a.FavoritesCount).Error
	return err
}

// 一个uid与多个aid之间的收藏关系
func (ar *articleRepo) GetIsFavorited(ctx context.Context, aids []uint, uid uint) (map[uint]bool, error) {
	var favorites []ArticleFavorite
	err := ar.data.db.Model(&ArticleFavorite{}).Where("user_id = ? AND article_id IN ?", uid, aids).Find(&favorites).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]bool)
	for _, aid := range aids {
		result[aid] = false
	}

	for _, favorite := range favorites {
		result[favorite.ArticleID] = true
	}

	return result, nil
}

// 查询文章
func (ar *articleRepo) ListArticlesByOptions(ctx context.Context, options *biz.ListOptions) ([]*biz.Article, error) {
	db := ar.data.db.Model(&Article{}).Preload("Author").Preload("Tags").Preload("Favorites")

	// 只返回当前用户相关的文章
	if options.CurrentUid > 0 && options.Tag == "" && options.Author == "" && options.FavoritedBy == "" {
		db = db.Joins("JOIN follows ON follows.following_id = articles.author_id").
			Where("follows.follower_id = ?", options.CurrentUid)
	} else {
		// 按标签过滤
		if options.Tag != "" {
			db = db.Joins("JOIN article_tags ON articles.id = article_tags.article_id").
				Joins("JOIN tags ON tags.id = article_tags.tag_id").
				Where("tags.name = ?", options.Tag)
		}
		// 按作者过滤
		if options.Author != "" {
			db = db.Joins("JOIN users ON users.id = articles.author_id").
				Where("users.username = ?", options.Author)
		}
		// 按被某用户收藏过滤
		if options.FavoritedBy != "" {
			db = db.Joins("JOIN article_favorites ON articles.id = article_favorites.article_id").
				Joins("JOIN users u2 ON u2.id = article_favorites.user_id").
				Where("u2.username = ?", options.FavoritedBy)
		}
	}

	// 分页
	if options.Limit > 0 {
		db = db.Limit(int(options.Limit))
	}
	if options.Offset > 0 {
		db = db.Offset(int(options.Offset))
	}

	// 执行查询
	db = db.Order("articles.created_at DESC")
	var articles []Article
	if err := db.Find(&articles).Error; err != nil {
		return nil, err
	}

	// 转换
	articleList := make([]*biz.Article, len(articles))
	for i, article := range articles {
		articleList[i] = convertArticle(article)
	}
	return articleList, nil
}

type commentRepo struct {
	data *Data
	log  *log.Helper
}

func NewCommentRepo(data *Data, logger log.Logger) biz.CommentRepo {
	return &commentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (cr *commentRepo) AddComment(ctx context.Context, c *biz.Comment) (*biz.Comment, error) {
	comment := Comment{
		ArticleID: c.ArticleID,
		AuthorID:  c.AuthorID,
		Body:      c.Body,
	}

	result := cr.data.db.Create(&comment)
	if result.Error != nil {
		return nil, result.Error
	}

	// 评论人的用户信息
	var user User
	if err := cr.data.db.Model(&User{}).Where("id = ?", c.AuthorID).First(&user).Error; err != nil {
		return nil, err
	}
	profile := &biz.ProfileResp{
		Username: user.Username,
		Bio:      user.Bio,
		Image:    user.Image,
	}

	return &biz.Comment{
		ID:        comment.ID,
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		Author:    profile,
	}, nil
}

func (cr *commentRepo) DeleteCommentByID(ctx context.Context, id uint) error {
	result := cr.data.db.Delete(&Comment{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("COMMENT_NOT_FOUND", "comment not found")
	}
	return nil
}

func (cr *commentRepo) GetCommentsByID(ctx context.Context, cid uint) ([]*biz.Comment, error) {
	var comments []Comment
	result := cr.data.db.Model(&Comment{}).Where("article_id = ?", cid).Preload("Author").Find(&comments)
	if result.Error != nil {
		return nil, result.Error
	}
	// 封装成biz.Comment
	commentList := make([]*biz.Comment, len(comments))
	for i, comment := range comments {
		commentList[i] = &biz.Comment{
			ID:        comment.ID,
			Body:      comment.Body,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
			Author: &biz.ProfileResp{
				Username: comment.Author.Username,
				Bio:      comment.Author.Bio,
				Image:    comment.Author.Image,
			},
		}
	}
	// todo: 登录用户和评论用户之间的follow关系
	return commentList, nil
}

type tagRepo struct {
	data *Data
	log  *log.Helper
}

func NewTagRepo(data *Data, logger log.Logger) biz.TagRepo {
	return &tagRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (tr *tagRepo) GetTags(ctx context.Context) ([]biz.Tag, error) {
	var tags []Tag
	err := tr.data.db.Find(&tags).Error
	if err != nil {
		return nil, err
	}
	// 转换
	tagList := make([]biz.Tag, len(tags))
	for i, tag := range tags {
		tagList[i] = biz.Tag(tag.Name)
	}
	return tagList, nil
}

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
	UserID    uint
	ArticleID uint `gorm:"index;constraint:OnDelete:CASCADE;"`
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

	// favorited count
	var fc int64
	article := convertArticle(a)
	e := ar.data.db.Model(&ArticleFavorite{}).Where("article_id = ?", a.ID).Count(&fc).Error
	if e != nil {
		return nil, e
	}
	article.FavoritesCount = uint32(fc)
	return article, nil
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

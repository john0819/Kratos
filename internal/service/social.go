package service

import (
	"context"

	v1 "kratos-realworld/api/realworld/v1"
	"kratos-realworld/internal/biz"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertArticle(a *biz.Article) *v1.SingleArticleResponse {
	return &v1.SingleArticleResponse{
		Article: &v1.Article{
			Slug:           a.Slug,
			Title:          a.Title,
			Description:    a.Description,
			Body:           a.Body,
			TagList:        a.TagList,
			CreatedAt:      timestamppb.New(a.CreatedAt),
			UpdatedAt:      timestamppb.New(a.UpdatedAt),
			Favorited:      a.Favorited,
			FavoritesCount: a.FavoritesCount,
			Author:         (*v1.Profile)(convertProfile(a.Author)),
		},
	}
}

func convertComment(c *biz.Comment) *v1.Comment {
	return &v1.Comment{
		Id:        uint32(c.ID),
		Body:      c.Body,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
		Author:    (*v1.Profile)(convertProfile(c.Author)),
	}
}

func (s *RealWorldService) ListArticles(ctx context.Context, req *v1.ListArticlesRequest) (*v1.MultipleArticleResponse, error) {
	// 通过配置项传递
	var opts []biz.ListOption
	if req.Limit > 0 {
		opts = append(opts, biz.WithLimit(int(req.Limit)))
	}
	if req.Offset > 0 {
		opts = append(opts, biz.WithOffset(int(req.Offset)))
	}
	if req.Tag != "" {
		opts = append(opts, biz.WithTag(req.Tag))
	}
	if req.Author != "" {
		opts = append(opts, biz.WithAuthor(req.Author))
	}
	if req.Favorited != "" {
		opts = append(opts, biz.WithFavoritedBy(req.Favorited))
	}

	a, err := s.uc.ListArticles(ctx, opts...)
	if err != nil {
		return nil, err
	}

	articles := make([]*v1.Article, 0)
	for _, article := range a {
		articles = append(articles, convertArticle(article).Article)
	}
	return &v1.MultipleArticleResponse{
		Articles:      articles,
		ArticlesCount: uint32(len(articles)),
	}, nil
}

func (s *RealWorldService) FeedArticles(ctx context.Context, in *v1.FeedArticlesRequest) (*v1.MultipleArticleResponse, error) {
	return &v1.MultipleArticleResponse{}, nil
}

func (s *RealWorldService) GetArticle(ctx context.Context, req *v1.GetArticleRequest) (*v1.SingleArticleResponse, error) {
	article, err := s.uc.GetArticle(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return convertArticle(article), nil
}

func (s *RealWorldService) CreateArticle(ctx context.Context, req *v1.CreateArticleRequest) (*v1.SingleArticleResponse, error) {
	article, err := s.uc.CreateArticle(ctx, &biz.Article{
		Title:       req.Article.Title,
		Description: req.Article.Description,
		Body:        req.Article.Body,
		TagList:     req.Article.TagList,
	})
	if err != nil {
		return nil, err
	}
	return convertArticle(article), nil
}

func (s *RealWorldService) UpdateArticle(ctx context.Context, req *v1.UpdateArticleRequest) (*v1.SingleArticleResponse, error) {
	article, err := s.uc.UpdateArticle(ctx, &biz.Article{
		Slug:        req.Slug,
		Title:       req.Article.Title,
		Description: req.Article.Description,
		Body:        req.Article.Body,
		TagList:     req.Article.TagList,
	})
	if err != nil {
		return nil, err
	}
	return convertArticle(article), nil
}

func (s *RealWorldService) DeleteArticle(ctx context.Context, req *v1.DeleteArticleRequest) (*v1.DeleteArticleResponse, error) {
	err := s.uc.DeleteArticle(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteArticleResponse{}, nil
}

func (s *RealWorldService) AddComment(ctx context.Context, req *v1.AddCommentRequest) (*v1.SingleCommentResponse, error) {
	comment, err := s.uc.AddComment(ctx, req.Slug, &biz.Comment{
		Body: req.Comment.Body,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SingleCommentResponse{
		Comment: convertComment(comment),
	}, nil
}

func (s *RealWorldService) GetComments(ctx context.Context, req *v1.GetCommentsRequest) (*v1.MultipleCommentResponse, error) {
	comments, err := s.uc.GetComments(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	// 转换
	commentList := make([]*v1.Comment, len(comments))
	for i, comment := range comments {
		commentList[i] = convertComment(comment)
	}
	return &v1.MultipleCommentResponse{
		Comments: commentList,
	}, nil
}

func (s *RealWorldService) DeleteComment(ctx context.Context, req *v1.DeleteCommentRequest) (*v1.DeleteCommentResponse, error) {
	err := s.uc.DeleteComment(ctx, req.Slug, uint(req.Id))
	if err != nil {
		return nil, err
	}
	return &v1.DeleteCommentResponse{
		Message: "delete comment success",
	}, nil
}

func (s *RealWorldService) FavoriteArticle(ctx context.Context, req *v1.FavoriteArticleRequest) (*v1.SingleArticleResponse, error) {
	article, err := s.uc.FavoriteArticle(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return convertArticle(article), nil
}

func (s *RealWorldService) UnfavoriteArticle(ctx context.Context, req *v1.UnfavoriteArticleRequest) (*v1.SingleArticleResponse, error) {
	article, err := s.uc.UnfavoriteArticle(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return convertArticle(article), nil
}

func (s *RealWorldService) GetTags(ctx context.Context, in *v1.GetTagsRequest) (*v1.TagsListResponse, error) {
	tags, err := s.uc.GetTags(ctx)
	if err != nil {
		return nil, err
	}
	tagList := make([]string, len(tags))
	for i, tag := range tags {
		tagList[i] = string(tag)
	}
	return &v1.TagsListResponse{
		Tags: tagList,
	}, nil
}

package service

import (
	"context"

	v1 "kratos-realworld/api/realworld/v1"
	"kratos-realworld/internal/biz"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertArticle(a *biz.Article) *v1.SingleArticleResponse {
	return &v1.SingleArticleResponse{
		Article: &v1.SingleArticleResponse_Article{
			Slug:           a.Slug,
			Title:          a.Title,
			Description:    a.Description,
			Body:           a.Body,
			TagList:        a.TagList,
			CreatedAt:      timestamppb.New(a.CreatedAt),
			UpdatedAt:      timestamppb.New(a.UpdatedAt),
			Favorited:      a.Favorited,
			FavoritesCount: a.FavoritesCount,
			Author:         (*v1.SingleArticleResponse_Author)(convertProfile(a.Author)),
		},
	}
}

func (s *RealWorldService) ListArticles(ctx context.Context, in *v1.ListArticlesRequest) (*v1.MultipleArticleResponse, error) {
	return &v1.MultipleArticleResponse{}, nil
}

func (s *RealWorldService) FeedArticles(ctx context.Context, in *v1.FeedArticlesRequest) (*v1.MultipleArticleResponse, error) {
	return &v1.MultipleArticleResponse{}, nil
}

func (s *RealWorldService) GetArticle(ctx context.Context, in *v1.GetArticleRequest) (*v1.SingleArticleResponse, error) {
	return &v1.SingleArticleResponse{}, nil
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

func (s *RealWorldService) UpdateArticle(ctx context.Context, in *v1.UpdateArticleRequest) (*v1.SingleArticleResponse, error) {
	return &v1.SingleArticleResponse{}, nil
}

func (s *RealWorldService) DeleteArticle(ctx context.Context, in *v1.DeleteArticleRequest) (*v1.DeleteArticleResponse, error) {
	return &v1.DeleteArticleResponse{}, nil
}

func (s *RealWorldService) AddComment(ctx context.Context, in *v1.AddCommentRequest) (*v1.SingleCommentResponse, error) {
	return &v1.SingleCommentResponse{}, nil
}

func (s *RealWorldService) GetComments(ctx context.Context, in *v1.GetCommentsRequest) (*v1.MultipleCommentResponse, error) {
	return &v1.MultipleCommentResponse{}, nil
}

func (s *RealWorldService) DeleteComment(ctx context.Context, in *v1.DeleteCommentRequest) (*v1.DeleteCommentResponse, error) {
	return &v1.DeleteCommentResponse{}, nil
}

func (s *RealWorldService) FavoriteArticle(ctx context.Context, in *v1.FavoriteArticleRequest) (*v1.SingleArticleResponse, error) {
	return &v1.SingleArticleResponse{}, nil
}

func (s *RealWorldService) UnfavoriteArticle(ctx context.Context, in *v1.UnfavoriteArticleRequest) (*v1.SingleArticleResponse, error) {
	return &v1.SingleArticleResponse{}, nil
}

func (s *RealWorldService) GetTags(ctx context.Context, in *v1.GetTagsRequest) (*v1.TagsListResponse, error) {
	return &v1.TagsListResponse{}, nil
}

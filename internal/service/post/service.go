// Package post implements post business logic.
package post

import (
	"context"
	"go-tweets/internal/config"
	"go-tweets/internal/dto"
	"go-tweets/internal/repository/comment"
	"go-tweets/internal/repository/post"
)

type PostService interface {
	CreatePost(ctx context.Context, model *dto.CreateOrUpdatePostRequest, userID int) (int, error)
	UpdatePost(ctx context.Context, model *dto.CreateOrUpdatePostRequest, postID, userID int) error
	DeletePost(ctx context.Context, postID, userID int) error
	LikeOrUnlikePost(ctx context.Context, postID, userID int) error
	DetailPost(ctx context.Context, postID int) (*dto.DetailPostResponse, error)
	GetAllPost(ctx context.Context, param *dto.GetAllPostRequest) (*dto.GetAllPostResponse, error)
}

type postService struct {
	cfg         *config.Config
	postRepo    post.PostRepository
	commentRepo comment.CommentRepository
}

func NewPostService(cfg *config.Config, postRepo post.PostRepository, commentRepo comment.CommentRepository) PostService {
	return &postService{
		cfg:         cfg,
		postRepo:    postRepo,
		commentRepo: commentRepo,
	}
}

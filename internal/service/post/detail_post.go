package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
)

func (s *postService) DetailPost(ctx context.Context, postID int) (*dto.DetailPostResponse, error) {

	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, apperror.Internal("failed to load post", err)
	}
	if post == nil {
		return nil, apperror.NotFound("post not found")
	}

	comments, err := s.commentRepo.GetCommentsByPostIDs(ctx, []int{postID})
	if err != nil {
		return nil, apperror.Internal("failed to load post comments", err)
	}
	commentMap := make([]dto.Comment, 0)
	for _, comment := range comments {
		commentMap = append(commentMap, dto.Comment{
			ID:        comment.ID,
			Username:  comment.Username,
			Content:   comment.Content,
			LikeCount: comment.LikeCount,
			CreatedAt: comment.CreatedAt.String(),
			UpdatedAt: comment.UpdatedAt.String(),
		})
	}
	return &dto.DetailPostResponse{
		ID:        post.ID,
		Username:  post.Username,
		Content:   post.Content,
		LikeCount: post.LikeCount,
		CreatedAt: post.CreatedAt.String(),
		UpdatedAt: post.UpdatedAt.String(),
		Comments:  commentMap,
	}, nil

}

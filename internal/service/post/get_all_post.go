package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"math"
)

func (s *postService) GetAllPost(ctx context.Context, param *dto.GetAllPostRequest) (*dto.GetAllPostResponse, error) {

	totalPost, err := s.postRepo.TotalPost(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to count posts", err)
	}

	offset := param.Limit * (param.Page - 1)
	posts, err := s.postRepo.GetAllPost(ctx, param, int(offset))
	if err != nil {
		return nil, apperror.Internal("failed to load posts", err)
	}

	postIDs := make([]int, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}
	comments, err := s.commentRepo.GetCommentsByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, apperror.Internal("failed to load post comments", err)
	}

	commentsMap := make(map[int][]dto.Comment)
	for _, comment := range comments {
		commentsMap[comment.PostID] = append(commentsMap[comment.PostID], dto.Comment{
			ID:        comment.ID,
			Username:  comment.Username,
			Content:   comment.Content,
			LikeCount: comment.LikeCount,
			CreatedAt: comment.CreatedAt.String(),
			UpdatedAt: comment.UpdatedAt.String(),
		})
	}

	var data []dto.DetailPostResponse
	for _, post := range posts {
		data = append(data, dto.DetailPostResponse{
			ID:        post.ID,
			Username:  post.Username,
			Title:     post.Title,
			Content:   post.Content,
			LikeCount: post.LikeCount,
			CreatedAt: post.CreatedAt.String(),
			UpdatedAt: post.UpdatedAt.String(),
			Comments:  commentsMap[post.ID],
		})
	}

	totalPage := int64(math.Ceil(float64(totalPost) / float64(param.Limit)))

	return &dto.GetAllPostResponse{
		TotalPage:   totalPage,
		CurrentPage: param.Page,
		Limit:       param.Limit,
		Data:        data,
	}, nil

}

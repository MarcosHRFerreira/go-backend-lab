package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
)

func (s *postService) DetailPost(ctx context.Context, postID int) (*dto.DetailPostResponse, error) {

	// Load the base post first because the detail response depends on its existence and ownership data.
	// Carrega primeiro o post base porque a resposta de detalhe depende da existencia e dos dados de autoria.
	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, apperror.Internal("failed to load post", err)
	}
	if post == nil {
		return nil, apperror.NotFound("post not found")
	}

	// Load related comments separately to keep repository queries focused on one aggregate at a time.
	// Carrega os comentarios relacionados separadamente para manter as queries do repository focadas em um agregado por vez.
	comments, err := s.commentRepo.GetCommentsByPostIDs(ctx, []int{postID})
	if err != nil {
		return nil, apperror.Internal("failed to load post comments", err)
	}
	// Map persistence models into transport DTOs before returning to the handler.
	// Mapeia os models de persistencia para DTOs de transporte antes de retornar ao handler.
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
	// Compose a single response object so the handler only serializes the final shape.
	// Compoe um unico objeto de resposta para que o handler apenas serializa a forma final.
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

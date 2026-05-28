package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/model"
	"time"
)

func (s *postService) LikeOrUnlikePost(ctx context.Context, postID, userID int) error {

	// Ensure the post exists before deciding whether the like should be inserted or removed.
	// Garante que o post existe antes de decidir se o like deve ser inserido ou removido.
	postExist, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return apperror.Internal("failed to load post", err)
	}
	if postExist == nil {
		return apperror.NotFound("tweet not found")
	}

	// Ask the repository for current state because toggle behavior depends on persisted data.
	// Consulta o repository pelo estado atual porque o comportamento de alternancia depende dos dados persistidos.
	isUserAlreadyLikePost, err := s.postRepo.IsUserAlreadyLikePost(ctx, postID, userID)
	if err != nil {
		return apperror.Internal("failed to check post like", err)
	}

	if isUserAlreadyLikePost {
		// Remove the like when the user has already reacted to the post.
		// Remove o like quando o usuario ja reagiu ao post.
		err := s.postRepo.DeleteLikePost(ctx, postID, userID)
		if err != nil {
			return apperror.Internal("failed to delete post like", err)
		}
	} else {
		// Create the like with fresh timestamps when this is the first reaction.
		// Cria o like com timestamps atuais quando esta e a primeira reacao.
		now := time.Now()
		err := s.postRepo.StoreLikePost(ctx, &model.PostLikeModel{
			UserID:    userID,
			PostID:    postID,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return apperror.Internal("failed to create post like", err)
		}

	}
	return nil
}

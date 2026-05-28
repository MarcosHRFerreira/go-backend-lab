package comment

import (
	"context"
	"go-tweets/internal/model"
)

func (r *commentRepository) StoreLikeComment(ctx context.Context, model *model.CommentLikeModel) error {
	// Insert the like row after the service decides the reaction is being added.
	// Insere a linha do like depois que o service decide que a reacao esta sendo adicionada.
	query := `INSERT INTO comment_likes ( comment_id, user_id, created_at, updated_at)
	VALUES(?,?,?,?) `

	_, err := r.db.ExecContext(ctx, query, model.CommentID, model.UserID, model.CreatedAt, model.UpdatedAt)
	return err

}

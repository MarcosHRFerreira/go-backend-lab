package comment

import (
	"context"
	"go-tweets/internal/model"
)

func (r *commentRepository) StoreComment(ctx context.Context, model *model.CommentModel) error {
	// Persist the comment after the service validates that the target post exists.
	// Persiste o comentario depois que o service valida que o post de destino existe.
	query := `INSERT INTO comments (post_id, user_id, content, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query, model.PostID, model.UserID, model.Content, model.CreatedAt, model.UpdatedAt)

	return err

}

package post

import (
	"context"
	"go-tweets/internal/model"
)

func (r *postRepository) StoreLikePost(ctx context.Context, model *model.PostLikeModel) error {
	// Persist a like row once the service decides the user is adding a new reaction.
	// Persiste uma linha de like quando o service decide que o usuario esta adicionando uma nova reacao.
	query := `INSERT INTO post_likes (post_id, user_id, created_at, updated_at)
	VALUES (?,?,?,?)`

	_, err := r.db.ExecContext(ctx, query, model.PostID, model.UserID, model.CreatedAt, model.UpdatedAt)

	return err

}

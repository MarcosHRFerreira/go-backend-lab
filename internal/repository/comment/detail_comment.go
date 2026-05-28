package comment

import (
	"context"
	"database/sql"
	"go-tweets/internal/model"
)

func (r *commentRepository) DetailComment(ctx context.Context, commentID int) (*model.CommentModel, error) {
	// Load the comment by identifier because like toggling must validate the target exists.
	// Carrega o comentario pelo identificador porque a alternancia de like precisa validar se o alvo existe.
	query := `SELECT id, post_id, user_id, content, created_at, updated_at
	FROM comments
	WHERE id = ?`
	// Scan into a typed model so upper layers stay independent from raw SQL fields.
	// Faz o scan em um model tipado para que as camadas acima fiquem independentes dos campos SQL brutos.
	row := r.db.QueryRowContext(ctx, query, commentID)
	var result model.CommentModel
	err := row.Scan(&result.ID, &result.PostID, &result.UserID, &result.Content, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Returning nil signals absence without turning it into a low-level database failure.
			// Retornar nil sinaliza ausencia sem transformar isso em uma falha de banco de baixo nivel.
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

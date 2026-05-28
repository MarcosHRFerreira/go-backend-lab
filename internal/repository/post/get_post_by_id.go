package post

import (
	"context"
	"database/sql"
	"go-tweets/internal/model"
)

func (r *postRepository) GetPostByID(ctx context.Context, postID int) (*model.PostWithUserModel, error) {
	// Join post, author, and like counters in one query because the detail view needs all of them together.
	// Faz o join de post, autor e contador de likes em uma unica query porque a visao de detalhe precisa de tudo junto.
	query := `SELECT p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at, u.username, COUNT(pl.id) as like_count
	FROM posts as p
	JOIN users as u ON p.user_id = u.id
	LEFT JOIN post_likes as pl ON pl.post_id = p.id
	WHERE p.id = ?
	AND p.delete_at IS NULL
	GROUP BY p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at, u.username`

	// Scan into a typed aggregate model so the service can apply business decisions without SQL details.
	// Faz o scan em um model agregado tipado para que o service aplique regras sem detalhes de SQL.
	row := r.db.QueryRowContext(ctx, query, postID)
	var result model.PostWithUserModel
	err := row.Scan(&result.ID, &result.Title, &result.Content, &result.UserID, &result.CreatedAt, &result.UpdatedAt, &result.Username, &result.LikeCount)
	if err != nil {
		if err == sql.ErrNoRows {
			// Returning nil, nil lets the service convert absence into a not-found response.
			// Retornar nil, nil permite que o service converta a ausencia em resposta de not found.
			return nil, nil
		}
		return nil, err
	}
	return &result, nil

}

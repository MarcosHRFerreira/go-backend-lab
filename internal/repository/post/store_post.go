package post

import (
	"context"
	"go-tweets/internal/model"
)

func (r *postRepository) StorePost(ctx context.Context, model *model.PostModel) (int, error) {
	// Insert the canonical post data after validation and ownership were enforced upstream.
	// Insere os dados canonicos do post depois que validacao e autoria foram garantidas acima.
	query := `INSERT INTO posts (user_id, title, content, created_at, updated_at) 
	VALUES (?,?,?,?,?)`

	result, err := r.db.ExecContext(ctx, query, model.UserID, model.Title, model.Content, model.CreatedAt, model.UpdatedAt)

	if err != nil {
		return 0, err
	}

	// Return the generated identifier so the service can expose it to the client.
	// Retorna o identificador gerado para que o service possa expor esse valor ao cliente.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil

}

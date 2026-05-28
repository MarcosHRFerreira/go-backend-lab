package post

import (
	"context"
	"go-tweets/internal/dto"
	"go-tweets/internal/model"
)

func (r *postRepository) GetAllPost(ctx context.Context, param *dto.GetAllPostRequest, offset int) ([]model.PostWithUserModel, error) {

	// The query denormalizes username and like count so the service can build the timeline response directly.
	// A query desnormaliza username e quantidade de likes para que o service possa montar a timeline diretamente.
	query := `SELECT
	p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at, u.username, COUNT(pl.id) as like_count
	FROM posts as p 
	JOIN users as u ON u.id = p.user_id
	LEFT JOIN post_likes as pl ON pl.post_id = p.id
	WHERE p.delete_at IS NULL
	GROUP BY p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at, u.username
	ORDER BY p.created_at DESC
	LIMIT ?
	OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, param.Limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan each row into a strongly typed model to keep SQL concerns inside the repository layer.
	// Faz o scan de cada linha em um model fortemente tipado para manter as preocupacoes de SQL dentro da camada de repository.
	result := make([]model.PostWithUserModel, 0)
	for rows.Next() {
		var data model.PostWithUserModel
		err := rows.Scan(&data.ID, &data.Title, &data.Content, &data.UserID, &data.CreatedAt, &data.UpdatedAt, &data.Username, &data.LikeCount)
		if err != nil {
			return nil, err
		}
		result = append(result, model.PostWithUserModel{
			ID:        data.ID,
			Title:     data.Title,
			Content:   data.Content,
			UserID:    data.UserID,
			CreatedAt: data.CreatedAt,
			UpdatedAt: data.UpdatedAt,
			Username:  data.Username,
			LikeCount: data.LikeCount,
		})
	}
	return result, nil

}

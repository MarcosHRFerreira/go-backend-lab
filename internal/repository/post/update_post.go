package post

import (
	"context"
	"errors"
	"go-tweets/internal/model"
)

func (r *postRepository) UpdatePost(ctx context.Context, model *model.PostModel, postID int) error {
	query := `UPDATE posts SET title = ?, content = ?, updated_at = ?
	WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, model.Title, model.Content, model.UpdatedAt, postID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("nothing to update")
	}
	return nil
}

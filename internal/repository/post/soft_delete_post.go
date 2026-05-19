package post

import (
	"context"
	"errors"
	"time"
)

func (r *postRepository) SoftDeletePost(ctx context.Context, postID int, now time.Time) error {
	query := `UPDATE posts SET delete_at = ?
	WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, now, postID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("nothing to update data")
	}
	return nil
}

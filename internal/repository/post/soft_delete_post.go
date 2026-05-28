package post

import (
	"context"
	"errors"
	"time"
)

func (r *postRepository) SoftDeletePost(ctx context.Context, postID int, now time.Time) error {
	// Mark the post as deleted instead of removing it physically to preserve historical data.
	// Marca o post como deletado em vez de removelo fisicamente para preservar dados historicos.
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
		// Report the absence of affected rows so upper layers can detect inconsistent state.
		// Informa a ausencia de linhas afetadas para que as camadas acima detectem estado inconsistente.
		return errors.New("nothing to update data")
	}
	return nil
}

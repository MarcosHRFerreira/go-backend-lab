package post

import (
	"context"
	"database/sql"
)

func (r *postRepository) IsUserAlreadyLikePost(ctx context.Context, postID, userID int) (bool, error) {
	// Probe for an existing like row because the service implements toggle semantics on top of this check.
	// Verifica se existe uma linha de like porque o service implementa a semantica de alternancia sobre essa checagem.
	query := `SELECT id FROM post_likes
	WHERE post_id = ?
	AND user_id = ?`

	row := r.db.QueryRowContext(ctx, query, postID, userID)
	var id int
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			// Absence means the next operation should create the like instead of removing it.
			// A ausencia significa que a proxima operacao deve criar o like em vez de remove-lo.
			return false, nil
		}
		return false, err
	}
	return true, nil

}

package comment

import (
	"context"
	"database/sql"
)

func (r *commentRepository) IsUserAlreadyLikeComment(ctx context.Context, commentID, userID int) (bool, error) {
	query := `SELECT id FROM comment_likes
	 WHERE comment_id = ? 
	 AND user_id = ?`

	row := r.db.QueryRowContext(ctx, query, commentID, userID)
	var id int
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil

}

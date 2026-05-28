package post

import (
	"context"
)

func (r *postRepository) DeleteLikePost(ctx context.Context, postID, userID int) error {
	// Delete the specific user reaction row when the toggle moves from liked to unliked.
	// Exclui a linha especifica da reacao do usuario quando a alternancia vai de liked para unliked.
	query := `DELETE FROM post_likes
	WHERE post_id = ?
	AND user_id = ?`

	_, err := r.db.ExecContext(ctx, query, postID, userID)

	return err

}

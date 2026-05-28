package user

import (
	"context"
	"errors"
)

func (r *userRepository) DeleteRefreshToken(ctx context.Context, userID int) error {
	// Remove the previous refresh token so each user keeps only the latest valid session token.
	// Remove o refresh token anterior para que cada usuario mantenha apenas o ultimo token de sessao valido.
	query := `DELETE FROM refresh_tokens
	WHERE user_id = ?`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Bubble up the absence so the service can decide whether the session state is inconsistent.
		// Propaga a ausencia para que o service decida se o estado da sessao esta inconsistente.
		return errors.New("nothing to delete")
	}
	return nil
}

package user

import (
	"context"
	"go-tweets/internal/model"
)

func (r *userRepository) StoreRefreshToken(ctx context.Context, model *model.RefreshTokenModel) error {
	// Persist the newly issued refresh token after the previous one has been invalidated.
	// Persiste o refresh token recem-emitido depois que o anterior foi invalidado.
	query := `INSERT INTO refresh_tokens (user_id, refresh_token, expired_at, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, model.UserID, model.RefreshToken, model.ExpiredAt, model.CreatedAt, model.UpdatedAt)

	return err
}

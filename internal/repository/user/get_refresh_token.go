package user

import (
	"context"
	"database/sql"
	"go-tweets/internal/model"
	"time"
)

func (r *userRepository) GetRefreshToken(ctx context.Context, userID int, now time.Time) (*model.RefreshTokenModel, error) {
	query := `SELECT id, user_id, refresh_token, expired_at
			  FROM refresh_tokens
			  WHERE user_id = ? AND expired_at > ?`

	row := r.db.QueryRowContext(ctx, query, userID, now)
	var result model.RefreshTokenModel
	err := row.Scan(&result.ID, &result.UserID, &result.RefreshToken, &result.ExpiredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil

}

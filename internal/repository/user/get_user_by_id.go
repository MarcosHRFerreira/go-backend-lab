package user

import (
	"context"
	"database/sql"
	"go-tweets/internal/model"
)

func (r *userRepository) GetUserByID(ctx context.Context, userID int) (*model.UserModel, error) {
	query := `SELECT id, username, email, created_at, updated_at
	FROM users WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, userID)
	var result model.UserModel
	err := row.Scan(&result.ID, &result.Username, &result.Email, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil

}

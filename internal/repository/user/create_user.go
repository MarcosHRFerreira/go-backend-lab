package user

import (
	"context"

	"go-tweets/internal/model"
)

func (r *userRepository) CreateUser(ctx context.Context, model *model.UserModel) (int, error) {
	query := `INSERT INTO users (email, username, password, created_at, updated_at)
	 VALUES (?,?,?,?,?)`
	result, err := r.db.ExecContext(ctx, query, model.Email, model.Username, model.Password, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(userID), nil

}

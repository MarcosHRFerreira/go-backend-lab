package user

import (
	"context"

	"go-tweets/internal/model"
)

func (r *userRepository) CreateUser(ctx context.Context, model *model.UserModel) (int, error) {
	// Persist the canonical user record after validation and hashing already happened upstream.
	// Persiste o registro canonico do usuario depois que validacao e hash ja aconteceram acima.
	query := `INSERT INTO users (email, username, password, created_at, updated_at)
	 VALUES (?,?,?,?,?)`
	result, err := r.db.ExecContext(ctx, query, model.Email, model.Username, model.Password, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return 0, err
	}

	// Return the generated primary key so the caller can build the HTTP response.
	// Retorna a chave primaria gerada para que o chamador monte a resposta HTTP.
	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(userID), nil

}

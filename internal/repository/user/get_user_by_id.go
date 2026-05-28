package user

import (
	"context"
	"database/sql"
	"go-tweets/internal/model"
)

func (r *userRepository) GetUserByID(ctx context.Context, userID int) (*model.UserModel, error) {
	// Reload the current user from the database when token flows need authoritative data.
	// Recarrega o usuario atual do banco quando fluxos de token precisam de dados autoritativos.
	query := `SELECT id, username, email, created_at, updated_at
	FROM users WHERE id = ?`

	// Keep the selected columns explicit to avoid leaking unnecessary sensitive fields.
	// Mantem as colunas selecionadas explicitas para evitar expor campos sensiveis desnecessarios.
	row := r.db.QueryRowContext(ctx, query, userID)
	var result model.UserModel
	err := row.Scan(&result.ID, &result.Username, &result.Email, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Absence is a valid business outcome, so return nil instead of forcing a repository error.
			// A ausencia e um resultado de negocio valido, entao retorna nil em vez de forcar erro de repository.
			return nil, nil
		}
		return nil, err
	}
	return &result, nil

}

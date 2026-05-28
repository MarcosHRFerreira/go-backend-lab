package user

import (
	"context"
	"database/sql"

	"go-tweets/internal/model"
)

func (r *userRepository) GetUserByEmailOrUsername(ctx context.Context, email, username string) (*model.UserModel, error) {
	// Query both unique fields because registration and login may search by either identity.
	// Consulta os dois campos unicos porque cadastro e login podem procurar por qualquer identidade.
	query := `SELECT id, username, email, password, created_at, updated_at
	FROM users
	WHERE email = ?
	or username = ?`

	// Scan straight into a typed model so callers receive a domain-friendly structure.
	// Faz o scan direto em um model tipado para que os chamadores recebam uma estrutura amigavel ao dominio.
	row := r.db.QueryRowContext(ctx, query, email, username)
	var result model.UserModel
	err := row.Scan(&result.ID, &result.Username, &result.Email, &result.Password, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Returning nil, nil lets the service decide whether absence means conflict or authentication failure.
			// Retornar nil, nil permite que o service decida se a ausencia significa conflito ou falha de autenticacao.
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

package user

import (
	"context"
	"database/sql"
	"go-tweets/internal/model"
	"time"
)

func (r *userRepository) GetRefreshToken(ctx context.Context, userID int, now time.Time) (*model.RefreshTokenModel, error) {
	// Load only non-expired refresh tokens because rotation must reject stale credentials.
	// Carrega apenas refresh tokens nao expirados porque a rotacao deve rejeitar credenciais vencidas.
	query := `SELECT id, user_id, refresh_token, expired_at
			  FROM refresh_tokens
			  WHERE user_id = ? AND expired_at > ?`

	// Map the row into a typed model so token comparison stays outside the repository.
	// Mapeia a linha para um model tipado para que a comparacao do token fique fora do repository.
	row := r.db.QueryRowContext(ctx, query, userID, now)
	var result model.RefreshTokenModel
	err := row.Scan(&result.ID, &result.UserID, &result.RefreshToken, &result.ExpiredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Missing active token is not exceptional; the service decides the semantic error.
			// A ausencia de token ativo nao e excepcional; o service decide o erro semantico.
			return nil, nil
		}
		return nil, err
	}

	return &result, nil

}

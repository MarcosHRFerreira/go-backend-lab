package comment

import (
	"context"
	"database/sql"
)

func (r *commentRepository) IsUserAlreadyLikeComment(ctx context.Context, commentID, userID int) (bool, error) {
	// Probe the join table to know whether the next action should insert or delete a reaction.
	// Consulta a tabela de associacao para saber se a proxima acao deve inserir ou excluir uma reacao.
	query := `SELECT id FROM comment_likes
	 WHERE comment_id = ? 
	 AND user_id = ?`

	row := r.db.QueryRowContext(ctx, query, commentID, userID)
	var id int
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			// No row means there is no current like registered for this user and comment.
			// Sem linha encontrada significa que nao existe like atual registrado para esse usuario e comentario.
			return false, nil
		}
		return false, err
	}
	return true, nil

}

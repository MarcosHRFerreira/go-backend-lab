package post

import (
	"context"
	"errors"
	"go-tweets/internal/model"
)

func (r *postRepository) UpdatePost(ctx context.Context, model *model.PostModel, postID int) error {
	// Update only mutable post fields because identity and authorship are immutable here.
	// Atualiza apenas os campos mutaveis do post porque identidade e autoria sao imutaveis aqui.
	query := `UPDATE posts SET title = ?, content = ?, updated_at = ?
	WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, model.Title, model.Content, model.UpdatedAt, postID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Signal the caller when nothing changed so upper layers can reason about stale targets.
		// Sinaliza ao chamador quando nada foi alterado para que as camadas acima tratem alvos obsoletos.
		return errors.New("nothing to update")
	}
	return nil
}

package post

import (
	"context"
	"go-tweets/internal/apperror"
	"go-tweets/internal/dto"
	"math"
)

func (s *postService) GetAllPost(ctx context.Context, param *dto.GetAllPostRequest) (*dto.GetAllPostResponse, error) {

	totalPost, err := s.postRepo.TotalPost(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to count posts", err)
	}

	// Pagination is computed in the service so the repository can stay focused on data access.
	// A paginacao e calculada no service para que o repository permaneca focado no acesso a dados.
	offset := param.Limit * (param.Page - 1)
	posts, err := s.postRepo.GetAllPost(ctx, param, int(offset))
	if err != nil {
		return nil, apperror.Internal("failed to load posts", err)
	}

	// Load comments in batch to avoid one query per post when building the response.
	// Carrega os comentarios em lote para evitar uma consulta por post ao montar a resposta.
	postIDs := make([]int, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}
	comments, err := s.commentRepo.GetCommentsByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, apperror.Internal("failed to load post comments", err)
	}

	// Group comments by post ID so the response can be assembled in a single pass.
	// Agrupa os comentarios por ID do post para que a resposta possa ser montada em uma unica passada.
	commentsMap := make(map[int][]dto.Comment)
	for _, comment := range comments {
		commentsMap[comment.PostID] = append(commentsMap[comment.PostID], dto.Comment{
			ID:        comment.ID,
			Username:  comment.Username,
			Content:   comment.Content,
			LikeCount: comment.LikeCount,
			CreatedAt: comment.CreatedAt.String(),
			UpdatedAt: comment.UpdatedAt.String(),
		})
	}

	// Convert repository models into DTOs that match the API response contract.
	// Converte os models do repository em DTOs que respeitam o contrato de resposta da API.
	var data []dto.DetailPostResponse
	for _, post := range posts {
		data = append(data, dto.DetailPostResponse{
			ID:        post.ID,
			Username:  post.Username,
			Title:     post.Title,
			Content:   post.Content,
			LikeCount: post.LikeCount,
			CreatedAt: post.CreatedAt.String(),
			UpdatedAt: post.UpdatedAt.String(),
			Comments:  commentsMap[post.ID],
		})
	}

	totalPage := int64(math.Ceil(float64(totalPost) / float64(param.Limit)))

	return &dto.GetAllPostResponse{
		TotalPage:   totalPage,
		CurrentPage: param.Page,
		Limit:       param.Limit,
		Data:        data,
	}, nil

}

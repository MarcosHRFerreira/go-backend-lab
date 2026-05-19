package dto

type (
	StoreCommentRequest struct {
		PostID  int    `json:"post_id" validate:"required"`
		Content string `json:"content" validate:"required"`
	}
)

type (
	LikeOrUnLikeCommentRequest struct {
		CommentID int `json:"comment_id" validate:"required"`
	}
)

// Package dto defines request and response payloads used by the API.
package dto

type (
	CreateOrUpdatePostRequest struct {
		Title   string `json:"title" validate:"required"`
		Content string `json:"content" validate:"required"`
	}
	CreateOrUpdatePostResponse struct {
		ID int `json:"id"`
	}
)

type LikeOrUnlikePostRequest struct {
	PostID int `json:"post_id" validate:"required"`
}

type (
	Comment struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Content   string `json:"content"`
		LikeCount int    `json:"like_count"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	DetailPostResponse struct {
		ID        int       `json:"id"`
		Username  string    `json:"username"`
		Title     string    `json:"title"`
		Content   string    `json:"content"`
		LikeCount int       `json:"like_count"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		Comments  []Comment `json:"comments"`
	}
)
type (
	GetAllPostRequest struct {
		Limit int64 `param:"limit"`
		Page  int64 `param:"page"`
	}
	GetAllPostResponse struct {
		TotalPage   int64                `json:"total_page"`
		CurrentPage int64                `json:"current_page"`
		Limit       int64                `json:"limit"`
		Data        []DetailPostResponse `json:"data"`
	}
)

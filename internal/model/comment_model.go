// Package model defines persistence models used across the application.
package model

import "time"

type (
	CommentModel struct {
		ID        int
		PostID    int
		UserID    int
		Username  string
		Content   string
		LikeCount int
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	CommentLikeModel struct {
		ID        int
		CommentID int
		UserID    int
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)

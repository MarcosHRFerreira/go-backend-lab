package model

import (
	"time"
)

type (
	PostModel struct {
		ID        int
		UserID    int
		Title     string
		Content   string
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt time.Time
	}

	PostLikeModel struct {
		ID        int
		PostID    int
		UserID    int
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	PostWithUserModel struct {
		ID        int
		UserID    int
		Username  string
		Title     string
		Content   string
		LikeCount int
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt time.Time
	}
)

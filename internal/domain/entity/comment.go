package entity

import "time"

type Comment struct {
	ID              int64
	PostID          int64
	OwnerID         int64
	ParentCommentID *int64
	Text            string
	Image           *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Status          bool
}

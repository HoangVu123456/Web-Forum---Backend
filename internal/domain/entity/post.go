package entity

import "time"

// Post represents a forum post
type Post struct {
	ID         int64
	OwnerID    int64
	CategoryID int64
	Headline   string
	Text       *string
	Image      *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Status     bool
}

package entity

import "time"

// Membership represents a user-category membership
type Membership struct {
	ID         int64
	CategoryID int64
	UserID     int64
	JoinedDate time.Time
}

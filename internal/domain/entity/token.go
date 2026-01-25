package entity

import "time"

// Token represents an authentication token for a user to use
type Token struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
}

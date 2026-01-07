package entity

import "time"

type Token struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
}

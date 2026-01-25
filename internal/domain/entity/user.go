package entity

import "time"

// User represents an account in the forum
type User struct {
	ID             int64
	Username       string
	Email          string
	Password       string
	ProfilePicture *string
	CreatedAt      time.Time
}

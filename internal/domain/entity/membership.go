package entity

import "time"

type Membership struct {
	ID         int64
	CategoryID int64
	UserID     int64
	JoinedDate time.Time
}

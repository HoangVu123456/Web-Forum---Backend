package entity

import "time"

type Notification struct {
	ID               int64
	OwnerID          int64
	ActorID          int64
	ComponentType    string
	ComponentID      int64
	NotificationType string
	Status           bool
	CreatedAt        time.Time
}

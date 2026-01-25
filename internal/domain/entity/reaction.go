package entity

// Reaction is a reaction on a post
type Reaction struct {
	ID             int64
	PostID         int64
	OwnerID        int64
	ReactionTypeID int64
}

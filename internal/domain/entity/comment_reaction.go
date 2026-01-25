package entity

// CommentReaction is a reaction on a comment or reply
type CommentReaction struct {
	ID             int64
	CommentID      int64
	OwnerID        int64
	ReactionTypeID int64
}

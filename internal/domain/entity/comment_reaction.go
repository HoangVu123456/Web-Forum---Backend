package entity

// CommentReaction is a reaction on a comment.
type CommentReaction struct {
	ID             int64
	CommentID      int64
	OwnerID        int64
	ReactionTypeID int64
}

package entity

// PostCategory links posts to categories.
type PostCategory struct {
	ID         int64
	PostID     int64
	CategoryID int64
}

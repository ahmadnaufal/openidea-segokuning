package post

import "time"

type CreatePostRequest struct {
	PostInHTML string   `json:"postInHtml" validate:"required,min=2,max=500"`
	Tags       []string `json:"tags" validate:"required"`

	UserID string
}

type ListPostsRequest struct {
	Limit     int      `query:"limit"`
	Offset    int      `query:"offset"`
	Search    string   `query:"search"`
	SearchTag []string `query:"searchTag"`

	UserID string
}

type AddCommentRequest struct {
	PostID  string `json:"postId" validate:"required"`
	Comment string `json:"comment" validate:"required,min=2,max=500"`

	UserID string
}

type Post struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	PostInHTML string    `db:"post_in_html"`
	CreatedAt  time.Time `db:"created_at"`

	Tags []PostTag
}

type PostComment struct {
	ID      string `db:"id"`
	UserID  string `db:"user_id"`
	PostID  string `db:"post_id"`
	Comment string `db:"comment"`
}

type PostTag struct {
	ID     int    `db:"id"`
	PostID string `db:"post_id"`
	Tag    string `db:"tag"`
}

type UserInPost struct {
	UserID        string    `db:"user_id"`
	Name          string    `db:"name"`
	ImageURL      string    `db:"image_url"`
	FriendCount   int       `db:"friend_count"`
	UserCreatedAt time.Time `db:"user_created_at"`
}

type PostDetail struct {
	// Post fields
	PostID        string    `db:"post_id"`
	PostInHTML    string    `db:"post_in_html"`
	PostCreatedAt time.Time `db:"post_created_at"`
	// Creator fields
	UserInPost
}

type CommentDetail struct {
	// Comment fields
	PostID  string `db:"post_id"`
	Comment string `db:"comment"`
	// Creator fields
	UserInPost
}

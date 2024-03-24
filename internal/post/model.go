package post

import (
	"database/sql"
	"errors"
	"time"
)

type CreatePostRequest struct {
	PostInHTML string   `json:"postInHtml" validate:"required,min=2,max=500"`
	Tags       []string `json:"tags" validate:"required,dive,min=1"`

	UserID string
}

type ListPostsRequest struct {
	Limit     uint     `query:"limit"`
	Offset    uint     `query:"offset"`
	Search    string   `query:"search"`
	SearchTag []string `query:"searchTag"`

	UserID string
	// RawQueries
	Queries map[string]string
}

// Validate is a function for additional validation related to query
func (r *ListPostsRequest) Validate() error {
	queries := r.Queries

	if val, ok := queries["limit"]; ok && val == "" {
		return errors.New("limit is empty")
	}

	if val, ok := queries["offset"]; ok && val == "" {
		return errors.New("offset is empty")
	}

	return nil
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
	UserID        string         `db:"user_id"`
	Name          string         `db:"name"`
	ImageURL      sql.NullString `db:"image_url"`
	FriendCount   int            `db:"friend_count"`
	UserCreatedAt time.Time      `db:"user_created_at"`
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

package friend

import (
	"database/sql"
	"time"
)

type FindFriendsRequest struct {
	Limit      int    `query:"limit"`
	Offset     int    `query:"offset"`
	SortBy     string `query:"sortBy"`
	OrderBy    string `query:"orderBy"`
	OnlyFriend bool   `query:"onlyFriend"`
	Search     string `query:"search"`

	UserID string
}

type AddFriendRequest struct {
	TargetUserID string `json:"userId"`
	UserID       string
}

type DeleteFriendRequest struct {
	TargetUserID string `json:"userId"`
	UserID       string
}

type UserFriend struct {
	UserID      string         `db:"user_id"`
	Name        string         `db:"name"`
	ImageURL    sql.NullString `db:"image_url"`
	FriendCount int            `db:"friend_count"`
	// CreatedAt is the user's register time, not when the friend request is created
	CreatedAt time.Time `db:"user_created_at"`
}

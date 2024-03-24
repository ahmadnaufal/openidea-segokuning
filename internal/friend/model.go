package friend

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

type FindFriendsRequest struct {
	Limit      uint   `query:"limit"`
	Offset     uint   `query:"offset"`
	SortBy     string `query:"sortBy"`
	OrderBy    string `query:"orderBy"`
	OnlyFriend bool   `query:"onlyFriend"`
	Search     string `query:"search"`

	UserID  string
	Queries map[string]string
}

var allowedSortByKey = map[string]bool{
	"friendcount": true,
	"createdat":   true,
}

var allowedOrderByKey = map[string]bool{
	"asc":  true,
	"desc": true,
}

func (r *FindFriendsRequest) Validate() error {
	queries := r.Queries

	if val, ok := queries["limit"]; ok && val == "" {
		return errors.New("limit is empty")
	}

	if val, ok := queries["offset"]; ok && val == "" {
		return errors.New("offset is empty")
	}

	if val, ok := queries["sortBy"]; ok && val == "" {
		return errors.New("sortBy is empty")
	}
	if _, found := allowedSortByKey[strings.ToLower(r.SortBy)]; r.SortBy != "" && !found {
		return errors.New("sortBy has invalid value")
	}

	if val, ok := queries["orderBy"]; ok && val == "" {
		return errors.New("orderBy is empty")
	}
	if _, found := allowedOrderByKey[strings.ToLower(r.OrderBy)]; r.SortBy != "" && !found {
		return errors.New("orderBy has invalid value")
	}

	if val, ok := queries["onlyFriend"]; ok && val == "" {
		return errors.New("onlyFriend is empty")
	}

	return nil
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

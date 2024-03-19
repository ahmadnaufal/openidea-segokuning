package friend

import "time"

type FriendResponse struct {
	UserID      string `json:"userId"`
	Name        string `json:"name"`
	ImageURL    string `json:"imageUrl"`
	FriendCount int    `json:"friendCount"`
	// CreatedAt is the user's register time, not when the friend request is created
	CreatedAt time.Time `json:"createdAt"`
}

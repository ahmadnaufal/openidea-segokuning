package post

import "time"

type CreatePostResponse struct {
	PostOnlyResponse
	PostID string `json:"postId"`
}

type PostOnlyResponse struct {
	PostInHTML string    `json:"postInHtml"`
	Tags       []string  `json:"tags"`
	CreatedAt  time.Time `json:"createdAt"`
}

type UserCreatorResponse struct {
	UserID      string    `json:"userId"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"imageUrl"`
	FriendCount int       `json:"friendCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CommentResponse struct {
	Comment string              `json:"comment"`
	Creator UserCreatorResponse `json:"creator"`
}

type PostDetailResponse struct {
	PostID   string              `json:"postId"`
	Post     PostOnlyResponse    `json:"post"`
	Comments []CommentResponse   `json:"comments"`
	Creator  UserCreatorResponse `json:"creator"`
}

type AddCommentResponse struct {
	PostID    string `json:"postId"`
	CommentID string `json:"commentId"`
	Comment   string `json:"comment"`
}

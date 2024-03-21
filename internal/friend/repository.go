package friend

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type FriendRepo struct {
	db *sqlx.DB
}

func NewFriendRepo(db *sqlx.DB) FriendRepo {
	return FriendRepo{db: db}
}

func (r *FriendRepo) IsUserFriendWith(ctx context.Context, userID, friendID string) (bool, error) {
	var isFriend bool

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM user_friends
			WHERE user_id_1=$1 AND user_id_2=$2
		) AS "exists"
	`

	err := r.db.GetContext(ctx, &isFriend, query, userID, friendID)
	if err != nil {
		return isFriend, err
	}

	return isFriend, nil
}

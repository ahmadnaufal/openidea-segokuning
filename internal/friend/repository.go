package friend

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func (r *FriendRepo) ListFriends(ctx context.Context, req FindFriendsRequest) ([]UserFriend, int, error) {
	var friends []UserFriend

	baseQuery := `
		SELECT
			u.id AS user_id,
			u.name AS name,
			u.image_url AS image_url,
			u.friend_count AS friend_count,
			u.created_at AS user_created_at
		FROM
			users u
		WHERE
			id != ?
		%s
	`

	// exclude the querying user
	args := []interface{}{req.UserID}

	filterQuery, filterArgs := getFilter(req)

	args = append(args, filterArgs...)

	queryWithFilter := fmt.Sprintf(baseQuery, filterQuery)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS temp", queryWithFilter)

	var count int
	err := r.db.GetContext(ctx, &count, sqlx.Rebind(sqlx.DOLLAR, countQuery), args...)
	if err != nil {
		return friends, count, err
	}

	orderQuery := getSortBy(req)
	limitQuery, limitArgs := getLimitAndOffset(req)
	args = append(args, limitArgs...)

	query := fmt.Sprintf("%s %s %s", queryWithFilter, orderQuery, limitQuery)

	err = r.db.SelectContext(ctx, &friends, sqlx.Rebind(sqlx.DOLLAR, query), args...)
	if err != nil {
		return friends, count, err
	}

	return friends, count, nil
}

func getFilter(req FindFriendsRequest) (string, []interface{}) {
	args := []interface{}{}
	filter := ""

	if req.OnlyFriend {
		filter += " AND u.id = ANY(SELECT user_id_2 FROM user_friends WHERE user_id_1 = ?)"
		args = append(args, req.UserID)
	}

	if req.Search != "" {
		filter += " AND (u.name ILIKE '%' || ? || '%' OR u.email ILIKE '%' || ? || '%' OR u.phone ILIKE '%' || ? || '%') "
		args = append(args, req.Search, req.Search, req.Search)
	}

	return filter, args
}

var sortKeyToColumnMap = map[string]string{
	"friendCount": "friend_count",
	"createdAt":   "created_at",
}

func getSortBy(req FindFriendsRequest) string {
	sortColumn := sortKeyToColumnMap[req.SortBy]
	if sortColumn == "" {
		sortColumn = "created_at"
	}

	sortOrdering := strings.ToUpper(req.OrderBy)
	if sortOrdering != "ASC" && sortOrdering != "DESC" {
		sortOrdering = "DESC"
	}

	query := fmt.Sprintf(`
		ORDER BY
			u.%s %s 
	`, sortColumn, sortOrdering)

	return query
}

func getLimitAndOffset(req FindFriendsRequest) (string, []interface{}) {
	// by default, set limit to 50
	query := "LIMIT ? OFFSET ?"

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	args := []interface{}{limit, offset}

	return query, args
}

func (r *FriendRepo) AddAsFriend(ctx context.Context, tx *sql.Tx, userID, friendID string) error {
	// 2 rows will be added:
	// 1. userID -> friendID (userID has friendID as friend)
	// 2. friendID -> userID (friendID has userID as friend)
	query := `
		INSERT INTO
			user_friends
			(user_id_1, user_id_2)
		VALUES
			($1, $2), ($2, $1)
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, userID, friendID)
	} else {
		_, err = r.db.ExecContext(ctx, query, userID, friendID)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *FriendRepo) DeleteFriend(ctx context.Context, tx *sql.Tx, userID, friendID string) error {
	// 2 rows will be removed:
	// 1. userID -> friendID (userID has friendID as friend)
	// 2. friendID -> userID (friendID has userID as friend)
	query := `
		DELETE FROM
			user_friends
		WHERE
			(user_id_1 = $1 AND user_id_2 = $2)
			OR
			(user_id_1 = $2 AND user_id_2 = $1)
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, userID, friendID)
	} else {
		_, err = r.db.ExecContext(ctx, query, userID, friendID)
	}
	if err != nil {
		return err
	}

	return nil
}

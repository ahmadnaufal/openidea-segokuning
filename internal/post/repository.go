package post

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type PostRepo struct {
	db *sqlx.DB
}

func NewPostRepo(db *sqlx.DB) PostRepo {
	return PostRepo{db: db}
}

func (r *PostRepo) CreatePost(ctx context.Context, tx *sql.Tx, post Post) error {
	query := `
		INSERT INTO
			posts
			(id, user_id, post_in_html)
		VALUES
			(:id, :user_id, :post_in_html)
	`

	updatedQuery, args, err := sqlx.Named(query, post)
	if err != nil {
		return err
	}

	// since we won't be using the returned data, leave it blank
	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	} else {
		_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *PostRepo) CreatePostTags(ctx context.Context, tx *sql.Tx, tags []PostTag) error {
	query := `
		INSERT INTO
			post_tags
			(post_id, tag)
		VALUES
			(:post_id, :tag)
	`

	updatedQuery, args, err := sqlx.Named(query, tags)
	if err != nil {
		return err
	}

	// since we won't be using the returned data, leave it blank
	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	} else {
		_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *PostRepo) GetPosts(ctx context.Context, req ListPostsRequest) ([]PostDetail, int, error) {
	var posts []PostDetail

	baseQuery := `
		SELECT
			p.id AS post_id,
			p.post_in_html AS post_in_html,
			p.created_at AS post_created_at,
			
			-- user fields
			p.user_id AS user_id,
			u.name AS name,
			u.image_url AS image_url,
			u.friend_count AS friend_count,
			u.created_at AS user_created_at
		FROM
			posts p
			INNER JOIN users u
			ON p.user_id = u.id
			INNER JOIN post_tags pt
			ON p.id = pt.post_id
		WHERE
			(
				p.user_id = ?
				OR p.user_id = ANY(
					SELECT
						user_id_2
					FROM
						user_friends
					WHERE
						user_id_1 = ?
				)
			)
		%s
	`

	args := []interface{}{req.UserID, req.UserID}

	filterQuery, filterArgs := getFilter(req)

	args = append(args, filterArgs...)

	queryWithFilter := fmt.Sprintf(baseQuery, filterQuery)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS temp", queryWithFilter)

	var count int
	err := r.db.GetContext(ctx, &count, sqlx.Rebind(sqlx.DOLLAR, countQuery), args...)
	if err != nil {
		return posts, count, err
	}

	orderQuery := getSortBy(req)
	limitQuery, limitArgs := getLimitAndOffset(req)
	args = append(args, limitArgs...)

	query := fmt.Sprintf("%s %s %s", queryWithFilter, orderQuery, limitQuery)

	err = r.db.SelectContext(ctx, &posts, sqlx.Rebind(sqlx.DOLLAR, query), args...)
	if err != nil {
		return posts, count, err
	}

	return posts, count, nil
}

func getFilter(req ListPostsRequest) (string, []interface{}) {
	args := []interface{}{}
	filter := ""

	// tags, bit complex
	if len(req.SearchTag) > 0 {
		placeholders := []string{}
		for _, tag := range req.SearchTag {
			args = append(args, tag)
			placeholders = append(placeholders, "?")
		}
		filter += fmt.Sprintf(" AND pt.tag IN (%s)", strings.Join(placeholders, ", "))
	}

	if req.Search != "" {
		filter += " AND p.post_in_html ILIKE '%' || ? || '%'"
		args = append(args, req.Search)
	}

	return filter, args
}

func getSortBy(_ ListPostsRequest) string {
	// hardcoded for now
	return `ORDER BY p.created_at DESC`
}

func getLimitAndOffset(req ListPostsRequest) (string, []interface{}) {
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

func (r *PostRepo) BulkGetPostComments(ctx context.Context, postIDs []string) (map[string][]CommentDetail, error) {
	var details []CommentDetail

	baseQuery := `
		SELECT
			pc.post_id AS post_id,
			pc.comment AS comment,
			
			-- user fields
			pc.user_id AS user_id,
			u.name AS name,
			u.image_url AS image_url,
			u.friend_count AS friend_count,
			u.created_at AS user_created_at
		FROM
			post_comments pc
			INNER JOIN users u
			ON pc.user_id = u.id
		WHERE
			post_id IN (?)
		ORDER BY
			pc.created_at DESC
	`

	updatedQuery, args, err := sqlx.In(baseQuery, postIDs)
	if err != nil {
		return nil, err
	}

	err = r.db.SelectContext(ctx, &details, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	if err != nil {
		return nil, err
	}

	postToCommentsMap := map[string][]CommentDetail{}
	for _, detail := range details {
		postID := detail.PostID
		postToCommentsMap[postID] = append(postToCommentsMap[postID], detail)
	}

	return postToCommentsMap, nil
}

func (r *PostRepo) BulkGetPostTags(ctx context.Context, postIDs []string) (map[string][]string, error) {
	var postTags []PostTag

	query := `
		SELECT
			id,
			post_id,
			tag
		FROM
			post_tags	
		WHERE
			post_id IN (?)
		ORDER BY
			post_id ASC, tag ASC	
	`

	updatedQuery, args, err := sqlx.In(query, postIDs)
	if err != nil {
		return nil, err
	}

	err = r.db.SelectContext(ctx, &postTags, r.db.Rebind(updatedQuery), args...)
	if err != nil {
		return nil, err
	}

	// group the fetched tags by each of post IDs
	postToTagMap := map[string][]string{}
	for _, tag := range postTags {
		productID := tag.PostID
		postToTagMap[productID] = append(postToTagMap[productID], tag.Tag)
	}

	return postToTagMap, nil
}

func (r *PostRepo) GetPostByID(ctx context.Context, postID string) (Post, error) {
	var post Post

	query := `
		SELECT
			id,
			user_id,
			post_in_html,
			created_at
		FROM
			posts
		WHERE
			id = $1
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &post, query, postID)
	if err != nil {
		return post, err
	}

	return post, nil
}

func (r *PostRepo) CreateComment(ctx context.Context, tx *sql.Tx, comment PostComment) error {
	query := `
		INSERT INTO
			post_comments
			(id, user_id, post_id, comment)
		VALUES
			(:id, :user_id, :post_id, :comment)
	`

	updatedQuery, args, err := sqlx.Named(query, comment)
	if err != nil {
		return err
	}

	// since we won't be using the returned data, leave it blank
	if tx != nil {
		_, err = tx.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	} else {
		_, err = r.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, updatedQuery), args...)
	}
	if err != nil {
		return err
	}

	return nil
}

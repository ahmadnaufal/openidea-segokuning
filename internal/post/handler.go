package post

import (
	"context"
	"database/sql"
	"html"
	"sync"
	"time"

	"github.com/ahmadnaufal/openidea-segokuning/internal/config"
	"github.com/ahmadnaufal/openidea-segokuning/internal/friend"
	"github.com/ahmadnaufal/openidea-segokuning/internal/model"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/jwt"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type postHandler struct {
	postRepo   *PostRepo
	txProvider *config.TransactionProvider
	friendRepo *friend.FriendRepo
}

type PostHandlerConfig struct {
	PostRepo   *PostRepo
	TxProvider *config.TransactionProvider
	FriendRepo *friend.FriendRepo
}

func NewPostHandler(cfg PostHandlerConfig) postHandler {
	return postHandler{
		postRepo:   cfg.PostRepo,
		txProvider: cfg.TxProvider,
		friendRepo: cfg.FriendRepo,
	}
}

func (h *postHandler) RegisterRoute(r *fiber.App, jwtProvider jwt.JWTProvider) {
	group := r.Group("/v1/post")
	authMiddleware := jwtProvider.Middleware()
	group.Use(authMiddleware)

	group.Get("/", h.ListPosts)
	group.Post("/", h.CreatePost)
	group.Post("/comment", h.AddComment)
}

func (h *postHandler) ListPosts(c *fiber.Ctx) error {
	var payload ListPostsRequest
	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return config.ErrRequestForbidden
	}
	payload.UserID = claims.UserID

	if err := c.QueryParser(&payload); err != nil {
		return errors.Wrap(config.ErrMalformedRequest, err.Error())
	}

	postResponses, meta, err := h.getPosts(c.Context(), payload)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "ok",
		Data:    postResponses,
		Meta:    &meta,
	})
}

func (h *postHandler) getPosts(ctx context.Context, payload ListPostsRequest) ([]PostDetailResponse, model.ResponseMeta, error) {
	var (
		responseMeta  model.ResponseMeta
		postResponses []PostDetailResponse = []PostDetailResponse{}
	)

	posts, count, err := h.postRepo.GetPosts(ctx, payload)
	if err != nil {
		return nil, responseMeta, errors.Wrap(err, "GetPosts error")
	}
	postIDs := []string{}
	for _, post := range posts {
		postIDs = append(postIDs, post.PostID)
	}

	// get post comments & get relevant tags (async)
	var (
		commentsMap   map[string][]CommentDetail
		tagsMap       map[string][]string
		getCommentErr error
		getTagErr     error
	)

	if len(postIDs) > 0 {
		wg := sync.WaitGroup{}

		wg.Add(2)
		go func() {
			defer wg.Done()
			commentsMap, getCommentErr = h.postRepo.BulkGetPostComments(ctx, postIDs)
		}()
		go func() {
			defer wg.Done()
			tagsMap, getTagErr = h.postRepo.BulkGetPostTags(ctx, postIDs)
		}()
		wg.Wait()

		if getCommentErr != nil {
			return nil, responseMeta, errors.Wrap(getCommentErr, "BulkGetPostComments error")
		}
		if getCommentErr != nil {
			return nil, responseMeta, errors.Wrap(getTagErr, "BulkGetPostTags error")
		}
	}

	// build response
	for _, post := range posts {
		// build comments
		comments := []CommentResponse{}
		postCommentDetails := commentsMap[post.PostID]
		for _, v := range postCommentDetails {
			comments = append(comments, CommentResponse{
				Comment: v.Comment,
				Creator: UserCreatorResponse{
					UserID:      v.UserID,
					Name:        v.Name,
					ImageURL:    v.ImageURL,
					FriendCount: v.FriendCount,
					CreatedAt:   v.UserCreatedAt,
				},
			})
		}

		postResponses = append(postResponses, PostDetailResponse{
			PostID: post.PostID,
			Post: PostOnlyResponse{
				PostInHTML: post.PostInHTML,
				Tags:       tagsMap[post.PostID],
				CreatedAt:  post.PostCreatedAt,
			},
			Comments: comments,
			Creator: UserCreatorResponse{
				UserID:      post.UserID,
				Name:        post.Name,
				ImageURL:    post.ImageURL,
				FriendCount: post.FriendCount,
				CreatedAt:   post.UserCreatedAt,
			},
		})
	}

	responseMeta.Limit = payload.Limit
	responseMeta.Offset = payload.Offset
	responseMeta.Total = count

	return postResponses, responseMeta, nil
}

func (h *postHandler) CreatePost(c *fiber.Ctx) error {
	var payload CreatePostRequest
	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return config.ErrRequestForbidden
	}
	payload.UserID = claims.UserID

	if err := c.BodyParser(&payload); err != nil {
		return errors.Wrap(config.ErrMalformedRequest, err.Error())
	}

	if err := validation.Validate(payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	post, err := h.createPostAndTags(c.Context(), payload)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(model.DataResponse{
		Message: "post created",
		Data: CreatePostResponse{
			PostID: post.ID,
			PostOnlyResponse: PostOnlyResponse{
				PostInHTML: post.PostInHTML,
				Tags:       payload.Tags,
				CreatedAt:  post.CreatedAt,
			},
		},
	})
}

func (h *postHandler) createPostAndTags(ctx context.Context, payload CreatePostRequest) (Post, error) {
	// create transaction
	tx, err := h.txProvider.NewTransaction(ctx)
	if err != nil {
		return Post{}, errors.Wrap(err, "NewTransaction error")
	}

	postID := uuid.NewString()
	post := Post{
		ID:         postID,
		UserID:     payload.UserID,
		PostInHTML: html.EscapeString(payload.PostInHTML),
		CreatedAt:  time.Now().UTC(),
	}

	err = h.postRepo.CreatePost(ctx, tx, post)
	if err != nil {
		return post, errors.Wrap(err, "CreatePost error")
	}

	// initialize post tags
	post.Tags = []PostTag{}
	for _, tag := range payload.Tags {
		post.Tags = append(post.Tags, PostTag{
			PostID: postID,
			Tag:    tag,
		})
	}

	err = h.postRepo.CreatePostTags(ctx, tx, post.Tags)
	if err != nil {
		return post, errors.Wrap(err, "CreatePost error")
	}

	err = tx.Commit()
	if err != nil {
		return post, errors.Wrap(err, "commit error")
	}

	return post, nil
}

func (h *postHandler) AddComment(c *fiber.Ctx) error {
	var payload AddCommentRequest
	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return config.ErrRequestForbidden
	}
	payload.UserID = claims.UserID

	if err := c.BodyParser(&payload); err != nil {
		return errors.Wrap(config.ErrMalformedRequest, err.Error())
	}

	if err := validation.Validate(payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	ctx := c.Context()
	// first, check if the post exists
	post, err := h.postRepo.GetPostByID(ctx, payload.PostID)
	if err != nil {
		if err == sql.ErrNoRows {
			return config.ErrPostNotFound
		}

		return errors.Wrap(err, "GetPostByID error")
	}

	// user can always comment on their own posts
	if post.UserID != payload.UserID {
		// check if the user is friend with the post creator
		isFriend, err := h.friendRepo.IsUserFriendWith(ctx, payload.UserID, post.UserID)
		if err != nil {
			return errors.Wrap(err, "IsUserFriendWith error")
		}
		if !isFriend {
			return config.ErrPostCreatorIsNotFriend
		}
	}

	// create the comment
	postComment := PostComment{
		ID:      uuid.NewString(),
		UserID:  payload.UserID,
		PostID:  payload.PostID,
		Comment: payload.Comment,
	}
	err = h.postRepo.CreateComment(ctx, nil, postComment)
	if err != nil {
		return errors.Wrap(err, "CreateComment error")
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "comment added",
		Data: AddCommentResponse{
			PostID:    postComment.PostID,
			CommentID: postComment.ID,
			Comment:   postComment.Comment,
		},
	})
}

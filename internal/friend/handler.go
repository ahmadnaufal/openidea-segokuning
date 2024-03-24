package friend

import (
	"context"
	"database/sql"

	"github.com/ahmadnaufal/openidea-segokuning/internal/config"
	"github.com/ahmadnaufal/openidea-segokuning/internal/model"
	"github.com/ahmadnaufal/openidea-segokuning/internal/user"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type friendHandler struct {
	userRepo   *user.UserRepo
	friendRepo *FriendRepo
	txProvider *config.TransactionProvider
}

type FriendHandlerConfig struct {
	UserRepo   *user.UserRepo
	FriendRepo *FriendRepo
	TxProvider *config.TransactionProvider
}

func NewFriendHandler(cfg FriendHandlerConfig) friendHandler {
	return friendHandler{
		userRepo:   cfg.UserRepo,
		friendRepo: cfg.FriendRepo,
		txProvider: cfg.TxProvider,
	}
}

func (h *friendHandler) RegisterRoute(r *fiber.App, jwtProvider jwt.JWTProvider) {
	group := r.Group("/v1/friend")
	authMiddleware := jwtProvider.Middleware()
	group.Use(authMiddleware)

	group.Get("/", h.FindFriends)
	group.Post("/", h.AddFriend)
	group.Delete("/", h.DeleteFriend)
}

func (h *friendHandler) FindFriends(c *fiber.Ctx) error {
	var payload FindFriendsRequest
	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return config.ErrRequestForbidden
	}
	payload.UserID = claims.UserID

	if err := c.QueryParser(&payload); err != nil {
		return errors.Wrap(config.ErrMalformedRequest, err.Error())
	}
	payload.Queries = c.Queries()
	if err := payload.Validate(); err != nil {
		return errors.Wrap(config.ErrMalformedRequest, err.Error())
	}

	userResponses, meta, err := h.getFriends(c.Context(), payload)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "ok",
		Data:    userResponses,
		Meta:    &meta,
	})
}

func (h *friendHandler) getFriends(ctx context.Context, payload FindFriendsRequest) ([]FriendResponse, model.ResponseMeta, error) {
	var (
		userResponses = []FriendResponse{}
		meta          model.ResponseMeta
	)

	users, count, err := h.friendRepo.ListFriends(ctx, payload)
	if err != nil {
		return userResponses, meta, errors.Wrap(err, "ListFriends error")
	}

	for _, user := range users {
		userResponses = append(userResponses, FriendResponse{
			UserID:      user.UserID,
			Name:        user.Name,
			ImageURL:    user.ImageURL.String,
			FriendCount: user.FriendCount,
			CreatedAt:   user.CreatedAt,
		})
	}

	meta.Limit = payload.Limit
	meta.Offset = payload.Offset
	meta.Total = uint(count)

	return userResponses, meta, nil
}

func (h *friendHandler) AddFriend(c *fiber.Ctx) error {
	var payload AddFriendRequest
	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return config.ErrRequestForbidden
	}
	payload.UserID = claims.UserID

	if err := c.BodyParser(&payload); err != nil {
		return errors.Wrap(config.ErrMalformedRequest, err.Error())
	}

	err = h.addFriend(c.Context(), payload)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "user added as friend",
	})
}

func (h *friendHandler) addFriend(ctx context.Context, payload AddFriendRequest) error {
	// check if user ID to be added as friend is empty
	if payload.TargetUserID == "" {
		return config.ErrMalformedRequest
	}

	// user cannot add themselves as friend
	if payload.TargetUserID == payload.UserID {
		return config.ErrSelfAddFriend
	}

	// check if the user exists
	targetFriend, err := h.userRepo.GetUserByID(ctx, payload.TargetUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return config.ErrUserNotFound
		}

		return errors.Wrap(err, "GetUserByID error")
	}

	// check if user already befriended
	isFriend, err := h.friendRepo.IsUserFriendWith(ctx, payload.UserID, targetFriend.ID)
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "IsUserFriendWith error")
	}
	if isFriend {
		return config.ErrFriendAlreadyAdded
	}

	// case: friend is not added yet. begin transaction
	// which will add each other as friend, then increment friendCount for each user by 1
	tx, err := h.txProvider.NewTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "NewTransaction error")
	}
	defer tx.Rollback()

	err = h.friendRepo.AddAsFriend(ctx, tx, payload.UserID, targetFriend.ID)
	if err != nil {
		return errors.Wrap(err, "AddAsFriend error")
	}

	// increment both friendCount counter
	err = h.userRepo.IncrementFriendCounter(ctx, tx, payload.UserID, targetFriend.ID)
	if err != nil {
		return errors.Wrap(err, "IncrementFriendCounter error")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Commit error")
	}

	return nil
}

func (h *friendHandler) DeleteFriend(c *fiber.Ctx) error {
	var payload DeleteFriendRequest
	claims, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return config.ErrRequestForbidden
	}
	payload.UserID = claims.UserID

	if err := c.BodyParser(&payload); err != nil {
		return errors.Wrap(config.ErrMalformedRequest, err.Error())
	}

	err = h.deleteFriend(c.Context(), payload)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "user removed from friend",
	})
}

func (h *friendHandler) deleteFriend(ctx context.Context, payload DeleteFriendRequest) error {
	// check if user ID to be added as friend is empty
	if payload.TargetUserID == "" {
		return config.ErrMalformedRequest
	}

	// user cannot add themselves from friend
	if payload.TargetUserID == payload.UserID {
		return config.ErrSelfAddFriend
	}

	// check if user already befriended
	isFriend, err := h.friendRepo.IsUserFriendWith(ctx, payload.UserID, payload.TargetUserID)
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "IsUserFriendWith error")
	}
	if !isFriend {
		return config.ErrUserIsNotAFriend
	}

	// case: friend is confirmed. begin transaction
	// which will remove each other from friend, then decrement friendCount for each user by 1
	tx, err := h.txProvider.NewTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "NewTransaction error")
	}
	defer tx.Rollback()

	err = h.friendRepo.DeleteFriend(ctx, tx, payload.UserID, payload.TargetUserID)
	if err != nil {
		return errors.Wrap(err, "DeleteFriend error")
	}

	// decrement both friendCount counter
	err = h.userRepo.DecrementFriendCounter(ctx, tx, payload.UserID, payload.TargetUserID)
	if err != nil {
		return errors.Wrap(err, "DecrementFriendCounter error")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Commit error")
	}

	return nil
}

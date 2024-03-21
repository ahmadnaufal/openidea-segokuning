package friend

import (
	"github.com/ahmadnaufal/openidea-segokuning/internal/model"
	"github.com/ahmadnaufal/openidea-segokuning/internal/user"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

type friendHandler struct {
	userRepo   *user.UserRepo
	friendRepo FriendRepo
}

type FriendHandlerConfig struct {
	UserRepo   *user.UserRepo
	FriendRepo *FriendRepo
}

func NewFriendHandler(cfg FriendHandlerConfig) friendHandler {
	return friendHandler{
		userRepo:   cfg.UserRepo,
		friendRepo: *cfg.FriendRepo,
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
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "ok",
		Data:    []FriendResponse{},
		Meta:    &model.ResponseMeta{},
	})
}

func (h *friendHandler) AddFriend(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "user added as friend",
	})
}

func (h *friendHandler) DeleteFriend(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "user removed from friend",
	})
}

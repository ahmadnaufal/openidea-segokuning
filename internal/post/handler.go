package post

import (
	"github.com/ahmadnaufal/openidea-segokuning/internal/model"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

type postHandler struct {
}

type PostHandlerConfig struct {
}

func NewPostHandler(cfg PostHandlerConfig) postHandler {
	return postHandler{}
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
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "ok",
		Data:    []PostDetailResponse{},
		Meta:    &model.ResponseMeta{},
	})
}

func (h *postHandler) CreatePost(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "post created",
		Data:    CreatePostResponse{},
	})
}

func (h *postHandler) AddComment(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "comment added",
		Data:    CommentResponse{},
	})
}

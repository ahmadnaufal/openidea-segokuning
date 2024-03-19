package user

import (
	"context"
	"time"

	"github.com/ahmadnaufal/openidea-segokuning/internal/model"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userHandler struct {
	userRepo    *UserRepo
	jwtProvider *jwt.JWTProvider
	saltCost    int
}

type UserHandlerConfig struct {
	UserRepo    *UserRepo
	JwtProvider *jwt.JWTProvider
	SaltCost    int
}

func NewUserHandler(cfg UserHandlerConfig) userHandler {
	return userHandler{
		userRepo:    cfg.UserRepo,
		jwtProvider: cfg.JwtProvider,
		saltCost:    cfg.SaltCost,
	}
}

func (h *userHandler) RegisterRoute(r *fiber.App, jwtProvider jwt.JWTProvider) {
	userGroup := r.Group("/v1/user")

	// middleware
	authMiddleware := jwtProvider.Middleware()

	userGroup.Post("/register", h.RegisterUser)
	userGroup.Post("/login", h.Authenticate)
	userGroup.Post("/link/:type", authMiddleware, h.LinkCredential)
	userGroup.Patch("/", authMiddleware, h.UpdateUser)
}

func (h *userHandler) RegisterUser(c *fiber.Ctx) error {
	h.createUser(c.Context(), RegisterUserRequest{})

	return c.Status(fiber.StatusCreated).JSON(model.DataResponse{
		Message: "User registered successfully",
		Data:    UserRegisterResponse{},
	})
}

func (h *userHandler) createUser(ctx context.Context, payload RegisterUserRequest) (User, error) {
	// hash the password first using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), h.saltCost)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:       uuid.NewString(),
		Name:     payload.Name,
		Password: string(hashedPassword),
	}
	err = h.userRepo.CreateUser(ctx, user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (h *userHandler) Authenticate(c *fiber.Ctx) error {
	if err := bcrypt.CompareHashAndPassword([]byte(""), []byte("")); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Wrong password entered for the username",
			Code:    "wrong_password",
		})
	}

	// generate JWT
	accessToken, err := h.generateAccessTokenFromUser(User{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "User registered successfully",
		Data: UserResponse{
			AccessToken: accessToken,
		},
	})
}

func (h *userHandler) generateAccessTokenFromUser(user User) (string, error) {
	claims := jwt.BuildJWTClaims(jwt.JWTUser{
		UserID: user.ID,
		Name:   user.Name,
	}, 8*time.Hour)

	accessToken, err := h.jwtProvider.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	return accessToken, err
}

func (h *userHandler) LinkCredential(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "credential updated successfully",
		Data:    UserResponse{},
	})
}

func (h *userHandler) UpdateUser(c *fiber.Ctx) error {

	return c.Status(fiber.StatusCreated).JSON(model.DataResponse{
		Message: "User updated successfully",
		Data:    UserResponse{},
	})
}

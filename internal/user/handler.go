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

var (
	UserRepoImpl *UserRepo
	JwtProvider  *jwt.JWTProvider
	SaltCost     int
)

func RegisterRoute(r *fiber.App) {
	r.Post("/v1/user/register", RegisterUser)
	r.Post("/v1/user/login", Authenticate)
}

func RegisterUser(c *fiber.Ctx) error {
	createUser(c.Context(), RegisterUserRequest{})

	return c.Status(fiber.StatusCreated).JSON(model.DataResponse{
		Message: "User registered successfully",
		Data:    UserAuthResponse{},
	})
}

func createUser(ctx context.Context, payload RegisterUserRequest) (User, error) {
	// hash the password first using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), SaltCost)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:       uuid.NewString(),
		Name:     payload.Name,
		Password: string(hashedPassword),
	}
	err = UserRepoImpl.CreateUser(ctx, user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func Authenticate(c *fiber.Ctx) error {
	if err := bcrypt.CompareHashAndPassword([]byte(""), []byte("")); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "Wrong password entered for the username",
			Code:    "wrong_password",
		})
	}

	// generate JWT
	accessToken, err := generateAccessTokenFromUser(User{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "User registered successfully",
		Data: UserAuthResponse{
			AccessToken: accessToken,
		},
	})
}

func generateAccessTokenFromUser(user User) (string, error) {
	claims := jwt.BuildJWTClaims(jwt.JWTUser{
		UserID: user.ID,
		Name:   user.Name,
	}, 8*time.Hour)

	accessToken, err := JwtProvider.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	return accessToken, err
}

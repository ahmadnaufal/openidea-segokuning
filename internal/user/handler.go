package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/ahmadnaufal/openidea-segokuning/internal/config"
	"github.com/ahmadnaufal/openidea-segokuning/internal/model"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/jwt"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
	var payload RegisterUserRequest
	if err := c.BodyParser(&payload); err != nil {
		return errors.Wrap(config.ErrMalformedRequest, err.Error())
	}

	if err := validation.Validate(payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// find existing user by credentials
	user, accessToken, err := h.createUser(c.Context(), payload)
	if err != nil {
		return errors.Wrap(err, "create user error")
	}

	return c.Status(fiber.StatusCreated).JSON(model.DataResponse{
		Message: "User registered successfully",
		Data: UserRegisterResponse{
			Email:       user.Email.String,
			Phone:       user.Phone.String,
			Name:        user.Name,
			AccessToken: accessToken,
		},
	})
}

func (h *userHandler) createUser(ctx context.Context, payload RegisterUserRequest) (User, string, error) {
	_, err := h.userRepo.GetUserByCredential(ctx, payload.CredentialType, payload.CredentialValue)
	if err != nil && err != sql.ErrNoRows {
		return User{}, "", errors.Wrap(err, "GetUserByCredential error")
	}
	if err == nil {
		// user already exists
		return User{}, "", config.ErrCredentialExists
	}

	// hash the password first using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), h.saltCost)
	if err != nil {
		return User{}, "", err
	}

	user := User{
		ID:       uuid.NewString(),
		Name:     payload.Name,
		Password: string(hashedPassword),
	}
	if payload.CredentialType == "email" {
		user.Email = sql.NullString{
			String: payload.CredentialValue,
			Valid:  true,
		}
	} else {
		user.Phone = sql.NullString{
			String: payload.CredentialValue,
			Valid:  true,
		}
	}
	err = h.userRepo.CreateUser(ctx, user)
	if err != nil {
		return user, "", err
	}

	// generate JWT
	accessToken, err := h.generateAccessTokenFromUser(user)
	if err != nil {
		return user, "", err
	}

	return user, accessToken, nil
}

func (h *userHandler) Authenticate(c *fiber.Ctx) error {
	var payload AuthenticateRequest
	if err := c.BodyParser(&payload); err != nil {
		return errors.Wrap(config.ErrMalformedRequest, err.Error())
	}

	if err := validation.Validate(payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, accessToken, err := h.authenticateUser(c.Context(), payload)
	if err != nil {
		return errors.Wrap(err, "create user error")
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "User registered successfully",
		Data: UserResponse{
			Email:       user.Email.String,
			Phone:       user.Phone.String,
			Name:        user.Name,
			AccessToken: accessToken,
		},
	})
}

func (h *userHandler) authenticateUser(ctx context.Context, payload AuthenticateRequest) (User, string, error) {
	user, err := h.userRepo.GetUserByCredential(ctx, payload.CredentialType, payload.CredentialValue)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, "", config.ErrUserNotFound
		}

		return user, "", errors.Wrap(err, "GetUserByCredential error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return user, "", config.ErrWrongPassword
	}

	// generate JWT
	accessToken, err := h.generateAccessTokenFromUser(user)
	if err != nil {
		return user, "", errors.Wrap(err, "generateAccessToken error")
	}

	return user, accessToken, nil
}

func (h *userHandler) generateAccessTokenFromUser(user User) (string, error) {
	claims := jwt.BuildJWTClaims(jwt.JWTUser{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email.String,
		Phone:  user.Phone.String,
	}, 8*time.Hour)

	accessToken, err := h.jwtProvider.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	return accessToken, err
}

func (h *userHandler) UpdateUser(c *fiber.Ctx) error {
	var payload UpdateUserRequest

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
	loggedInUser, err := h.updateUser(ctx, payload)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "user updated successfully",
		Data: UserResponse{
			Phone: loggedInUser.Phone.String,
			Email: loggedInUser.Email.String,
			Name:  loggedInUser.Name,
		},
	})
}

func (h *userHandler) updateUser(ctx context.Context, payload UpdateUserRequest) (User, error) {
	loggedInUser, err := h.userRepo.GetUserByID(ctx, payload.UserID)
	if err != nil {
		return User{}, errors.Wrap(err, "getuserByID error")
	}

	// do update credential
	loggedInUser.ImageURL = sql.NullString{
		String: payload.ImageURL,
		Valid:  true,
	}
	loggedInUser.Name = payload.Name
	err = h.userRepo.UpdateUser(ctx, nil, loggedInUser)
	if err != nil && err != sql.ErrNoRows {
		return loggedInUser, errors.Wrap(err, "UpdateUser error")
	}

	return loggedInUser, nil
}

func (h *userHandler) LinkCredential(c *fiber.Ctx) error {
	var payload LinkCredentialRequest
	payload.CredentialType = c.Params("type")
	if payload.CredentialType != "email" && payload.CredentialType != "phone" {
		return fiber.ErrNotFound
	}

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
	loggedInUser, err := h.updateCredential(ctx, payload)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "credential updated successfully",
		Data: UserResponse{
			Phone: loggedInUser.Phone.String,
			Email: loggedInUser.Email.String,
			Name:  loggedInUser.Name,
		},
	})
}

func (h *userHandler) updateCredential(ctx context.Context, payload LinkCredentialRequest) (User, error) {
	loggedInUser, err := h.userRepo.GetUserByID(ctx, payload.UserID)
	if err != nil {
		return User{}, errors.Wrap(err, "getuserByID error")
	}

	var credentialValue string
	if payload.CredentialType == "email" {
		// if credentialType is email, but user already have email registered
		// do not allow user to update email from this API
		if loggedInUser.Email.String != "" {
			return loggedInUser, config.ErrCannotChangeCredential
		}

		credentialValue = payload.Email
	} else {
		// if credentialType is phone, but user already have phone registered
		// do not allow user to update email from this API
		if loggedInUser.Phone.String != "" {
			return loggedInUser, config.ErrCannotChangeCredential
		}

		credentialValue = payload.Phone
	}

	// check for existing user which already have the credential
	// since we already guaranteed user having email/phone cannot use this api to update their existing email/phone
	// the user fetched here will be another user
	_, err = h.userRepo.GetUserByCredential(ctx, payload.CredentialType, credentialValue)
	if err != nil && err != sql.ErrNoRows {
		return loggedInUser, errors.Wrap(err, "GetUserByCredential error")
	}
	if err == nil {
		// user already exists
		return loggedInUser, config.ErrCredentialExists
	}

	// do update credential
	if payload.CredentialType == "email" {
		loggedInUser.Email = sql.NullString{
			String: credentialValue,
			Valid:  true,
		}
	} else {
		loggedInUser.Phone = sql.NullString{
			String: credentialValue,
			Valid:  true,
		}
	}
	err = h.userRepo.UpdateUser(ctx, nil, loggedInUser)
	if err != nil && err != sql.ErrNoRows {
		return loggedInUser, errors.Wrap(err, "UpdateUser error")
	}

	return loggedInUser, nil
}

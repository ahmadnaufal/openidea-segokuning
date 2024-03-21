package config

import (
	"net/http"

	"github.com/ahmadnaufal/openidea-segokuning/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

var (
	ErrMalformedRequest       = fiber.NewError(http.StatusBadRequest, "request malformed")
	ErrCredentialExists       = fiber.NewError(http.StatusConflict, "credential already used")
	ErrUserNotFound           = fiber.NewError(http.StatusNotFound, "user with the specified credential not found")
	ErrWrongPassword          = fiber.NewError(http.StatusBadRequest, "wrong password entered")
	ErrRequestForbidden       = fiber.NewError(http.StatusForbidden, "request forbidden")
	ErrCannotChangeCredential = fiber.NewError(http.StatusBadRequest, "cannot change email or phone from link email/phone API")
)

func DefaultErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Status code defaults to 500
		code := fiber.StatusInternalServerError
		message := "internal server error"

		// Retrieve the custom status code & message if it's a *fiber.Error
		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
			message = e.Message
		}

		// Return status code with error message
		return c.Status(code).JSON(model.ErrorResponse{
			Message: message,
		})
	}
}

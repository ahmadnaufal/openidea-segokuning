package image

import (
	"strings"

	"github.com/ahmadnaufal/openidea-segokuning/internal/model"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/jwt"
	"github.com/ahmadnaufal/openidea-segokuning/pkg/s3"
	"github.com/gofiber/fiber/v2"
)

type imageHandler struct {
	s3Provider *s3.S3Provider
}

func NewImageHandler(s3prov *s3.S3Provider) imageHandler {
	return imageHandler{
		s3Provider: s3prov,
	}
}

func (h *imageHandler) RegisterRoute(r *fiber.App, jwtProvider jwt.JWTProvider) {
	imageGroup := r.Group("/v1/image")
	authMiddleware := jwtProvider.Middleware()

	imageGroup.Post("/", authMiddleware, h.UploadImage)
}

func (h *imageHandler) UploadImage(c *fiber.Ctx) error {
	// check for credentials
	_, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
		})
	}

	fileReader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "uploaded file was invalid",
		})
	}

	// check file size & extension
	fileSize := fileReader.Size
	if fileSize < 10*1024 || fileSize > 2*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "invalid file size",
		})
	}

	// check extension
	fp := strings.Split(fileReader.Filename, ".")
	if len(fp) < 2 || (fp[len(fp)-1] != "jpg" && fp[len(fp)-1] != "jpeg") {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "invalid file extension",
		})
	}

	imgUrl, err := h.s3Provider.UploadImage(c.Context(), fileReader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
		})
	}

	return c.Status(fiber.StatusOK).JSON(model.DataResponse{
		Message: "File uploaded successfully",
		Data: ImageUploadResponse{
			ImageURL: imgUrl,
		},
	})
}

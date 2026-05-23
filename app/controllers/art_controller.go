package controllers

import (
	"Continuous_b/app/services"

	"github.com/gofiber/fiber/v2"
)

// ArtController handles the string art processing endpoint.
type ArtController struct {
	imageService    *services.ImageService
	stringArtService *services.StringArtService
}

// NewArtController creates a new ArtController with injected services.
func NewArtController() *ArtController {
	return &ArtController{
		imageService:    services.NewImageService(),
		stringArtService: services.NewStringArtService(),
	}
}

// Process handles POST /api/v1/process
// Accepts a multipart/form-data with field "image"
// Returns JSON: { "steps": [...], "total_pegs": 200, "total_lines": 3000 }
func (a *ArtController) Process(c *fiber.Ctx) error {
	// 1. Read uploaded file
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "image field is required (multipart/form-data)",
		})
	}

	// Validate file size (max 10 MB)
	const maxSize = 10 * 1024 * 1024
	if fileHeader.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "image too large, max 10 MB",
		})
	}

	// 2. Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open uploaded file",
		})
	}
	defer file.Close()

	// 3. Preprocess image (grayscale + contrast + resize)
	processed, err := a.imageService.PrepareFromFile(file)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "failed to process image: " + err.Error(),
		})
	}

	// 4. Run greedy string art algorithm
	result := a.stringArtService.Compute(processed)

	// 5. Return the peg sequence
	return c.JSON(result)
}

package services

import (
	"fmt"
	"image"
	"image/color"
	"mime/multipart"

	"github.com/disintegration/imaging"
)

// ImageService handles image preprocessing for string art generation.
type ImageService struct{}

// NewImageService creates a new ImageService.
func NewImageService() *ImageService {
	return &ImageService{}
}

// ProcessedImage holds the grayscale pixel matrix ready for the algorithm.
type ProcessedImage struct {
	Width  int
	Height int
	// Pixels[y][x] = 0 (black) .. 255 (white)
	Pixels [][]uint8
}

// PrepareFromFile reads a multipart file, converts it to grayscale,
// enhances contrast, and returns a pixel matrix.
func (s *ImageService) PrepareFromFile(file multipart.File) (*ProcessedImage, error) {
	// Decode image (supports JPEG, PNG, GIF, BMP, TIFF, WEBP via imaging)
	src, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 1. Resize to a consistent square for the circular template (600x600)
	resized := imaging.Fit(src, 600, 600, imaging.Lanczos)

	// 2. Convert to grayscale
	gray := imaging.Grayscale(resized)

	// 3. Enhance contrast naturally so dark lines stand out cleanly.
	//    30 is a well-balanced value that preserves skin gradients.
	contrasted := imaging.AdjustContrast(gray, 30)

	// 4. Sharpen slightly to recover edge detail lost in resize.
	sharpened := imaging.Sharpen(contrasted, 1.0)

	bounds := sharpened.Bounds()
	w := bounds.Max.X
	h := bounds.Max.Y

	// Build pixel matrix
	pixels := make([][]uint8, h)
	for y := 0; y < h; y++ {
		pixels[y] = make([]uint8, w)
		for x := 0; x < w; x++ {
			c := sharpened.At(x+bounds.Min.X, y+bounds.Min.Y)
			r, _, _, _ := color.GrayModel.Convert(c).RGBA()
			// r is 16-bit (0..65535), convert to 8-bit
			pixels[y][x] = uint8(r >> 8)
		}
	}

	return &ProcessedImage{
		Width:  w,
		Height: h,
		Pixels: pixels,
	}, nil
}

// Ensure image.Image is accessible for bounds checks
var _ image.Image = (*image.NRGBA)(nil)

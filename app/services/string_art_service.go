package services

import (
	"math"
)

const (
	// NumPegs is the number of nails distributed evenly on the circular frame.
	NumPegs = 200

	// NumLines is the total number of thread segments to compute.
	NumLines = 3000

	// MinPegDistance prevents drawing very short lines between adjacent pegs.
	MinPegDistance = 25

	// LineWeight controls how much darkness is subtracted after drawing each line.
	// 40.0 is the original balanced weight.
	LineWeight = 40.0
)

// StringArtService computes the greedy string art sequence.
type StringArtService struct{}

// NewStringArtService creates a new StringArtService.
func NewStringArtService() *StringArtService {
	return &StringArtService{}
}

// ComputeResult holds the output of the algorithm.
type ComputeResult struct {
	Steps      []int `json:"steps"`
	TotalPegs  int   `json:"total_pegs"`
	TotalLines int   `json:"total_lines"`
}

// pegCoords returns (x, y) pixel coordinates for peg index i
// inscribed in a circle centred in the given image dimensions.
func pegCoords(i, numPegs, width, height int) (float64, float64) {
	cx := float64(width) / 2.0
	cy := float64(height) / 2.0
	r := math.Min(cx, cy) * 0.97 // small inset so pegs aren't clipped
	angle := 2.0 * math.Pi * float64(i) / float64(numPegs)
	return cx + r*math.Cos(angle), cy + r*math.Sin(angle)
}

// bresenhamLine returns all (x, y) integer pixel positions along a line
// from (x0,y0) to (x1,y1) using Bresenham's algorithm.
func bresenhamLine(x0, y0, x1, y1 int) [][2]int {
	points := make([][2]int, 0, max(abs(x1-x0), abs(y1-y0))+1)

	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		points = append(points, [2]int{x0, y0})
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
	return points
}

// getPixelScore converts a pixel value (0=black, 255=white) to a darkness score.
// It uses a continuous power function to naturally prioritize dark features (eyes, hair)
// while allowing soft threads to pass through light areas (skin) without creating artificial white blocks.
func getPixelScore(v uint8) float64 {
	d := float64(255 - v)
	// Power 1.6 gives a very natural contrast balance
	return math.Pow(d, 1.6)
}

// lineDarkness computes the average darkness score of pixels along a line.
// Returns -1 if the line has no valid pixels.
func lineDarkness(pixels [][]uint8, w, h int, pts [][2]int) float64 {
	if len(pts) == 0 {
		return -1
	}
	total := 0.0
	count := 0
	for _, pt := range pts {
		x, y := pt[0], pt[1]
		if x >= 0 && x < w && y >= 0 && y < h {
			total += getPixelScore(pixels[y][x])
			count++
		}
	}
	if count == 0 {
		return -1
	}
	return total / float64(count)
}

// safeSubtract lightens a single pixel by adding the given amount, clamping to 255.
func safeSubtract(pixels [][]uint8, w, h, x, y int, amt float64) {
	if x >= 0 && x < w && y >= 0 && y < h {
		v := float64(pixels[y][x]) + amt
		if v > 255 {
			v = 255
		}
		pixels[y][x] = uint8(v)
	}
}

// subtractLine lightens the pixels along a drawn thread so the algorithm
// doesn't keep drawing on the same spot. It uses a very soft 3-pixel wide brush
// (12% weight on neighbors) to keep lines crisp and detailed while preventing hard banding.
func subtractLine(pixels [][]uint8, w, h int, pts [][2]int, weight float64) {
	for _, pt := range pts {
		x, y := pt[0], pt[1]
		
		// Center pixel gets full weight
		safeSubtract(pixels, w, h, x, y, weight)
		
		// Neighbors get very soft weight (12% of center) to prevent banding without blurring
		safeSubtract(pixels, w, h, x+1, y, weight*0.12)
		safeSubtract(pixels, w, h, x-1, y, weight*0.12)
		safeSubtract(pixels, w, h, x, y+1, weight*0.12)
		safeSubtract(pixels, w, h, x, y-1, weight*0.12)
	}
}

// Compute runs the greedy string art algorithm on the preprocessed image.
// Returns the sequence of peg indices to wind the thread around.
func (s *StringArtService) Compute(img *ProcessedImage) *ComputeResult {
	w, h := img.Width, img.Height

	// Deep-copy pixels so we modify a local canvas, not the original
	canvas := make([][]uint8, h)
	for y := 0; y < h; y++ {
		canvas[y] = make([]uint8, w)
		copy(canvas[y], img.Pixels[y])
	}

	// ── Circular Masking ─────────────────────────────────────────────────────
	// Set all pixels outside the hoop boundary to pure white (255).
	cx := float64(w) / 2.0
	cy := float64(h) / 2.0
	maskR := math.Min(cx, cy) * 0.98
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			if math.Sqrt(dx*dx+dy*dy) > maskR {
				canvas[y][x] = 255
			}
		}
	}

	// Precompute peg coordinates
	pegX := make([]int, NumPegs)
	pegY := make([]int, NumPegs)
	for i := 0; i < NumPegs; i++ {
		fx, fy := pegCoords(i, NumPegs, w, h)
		pegX[i] = int(math.Round(fx))
		pegY[i] = int(math.Round(fy))
	}

	steps := make([]int, 0, NumLines+1)
	current := 0
	steps = append(steps, current)

	for line := 0; line < NumLines; line++ {
		bestPeg := -1
		bestScore := -1.0

		for next := 0; next < NumPegs; next++ {
			if next == current {
				continue
			}
			// Enforce minimum angular distance to avoid near-zero lines
			dist := next - current
			if dist < 0 {
				dist = -dist
			}
			if dist < MinPegDistance || dist > NumPegs-MinPegDistance {
				continue
			}

			pts := bresenhamLine(pegX[current], pegY[current], pegX[next], pegY[next])
			score := lineDarkness(canvas, w, h, pts)
			if score > bestScore {
				bestScore = score
				bestPeg = next
			}
		}

		if bestPeg == -1 {
			// Fallback: pick opposite peg
			bestPeg = (current + NumPegs/2) % NumPegs
		}

		// Draw the chosen line (subtract from canvas)
		pts := bresenhamLine(pegX[current], pegY[current], pegX[bestPeg], pegY[bestPeg])
		subtractLine(canvas, w, h, pts, LineWeight)

		steps = append(steps, bestPeg)
		current = bestPeg
	}

	return &ComputeResult{
		Steps:      steps,
		TotalPegs:  NumPegs,
		TotalLines: NumLines,
	}
}

// ---- helpers ----

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

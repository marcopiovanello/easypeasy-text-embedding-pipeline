package utils

import (
	"image"
	"image/draw"
)

func ToGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	grayscale := image.NewGray(bounds)

	draw.Draw(grayscale, bounds, img, bounds.Min, draw.Src)

	return grayscale
}

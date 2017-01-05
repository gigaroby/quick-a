package imageutil

import (
	"image"
	"image/color"
)

var (
	// AlphaAsWhite creates a color model that will convert any pixel with
	// alpha=0 and replace it with a white pixel with no transparency
	AlphaAsWhite = color.ModelFunc(alphaAsWhite)
)

func alphaAsWhite(c color.Color) color.Color {
	if _, ok := c.(color.Gray); ok {
		return c
	}
	_, _, _, a := c.RGBA()
	if a == 0 {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	return c
}

// ConvertTo returns an image that applies the provided model to every pixel.
func ConvertTo(original image.Image, model color.Model) image.Image {
	return &converted{
		original: original,
		model:    model,
	}
}

// converted implements image.Image
type converted struct {
	// original holds the original image
	original image.Image
	// model holds the color model that will be used for the conversion
	model color.Model
}

func (c *converted) ColorModel() color.Model {
	return c.model
}

func (c *converted) Bounds() image.Rectangle {
	return c.original.Bounds()
}

func (c *converted) At(x, y int) color.Color {
	return c.model.Convert(c.original.At(x, y))
}

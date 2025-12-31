package ircimage

import (
	"errors"
	"image"
)

func ConvertImageWithHexColors(img *image.NRGBA) (HalfBlockImage, error) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	if h%2 != 0 {
		return nil, errors.New("height must be a multiple of 2")
	}

	out := make([][]HalfBlockChar, h/2)

	for y := 0; y < h; y += 2 {
		out[y/2] = make([]HalfBlockChar, w)
		for x := range w {
			out[y/2][x] = &HalfBlockHexColor{
				Top: img.NRGBAAt(x, y), Bottom: img.NRGBAAt(x, y+1),
			}
		}
	}

	return HalfBlockImage(out), nil
}

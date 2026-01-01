package ircimage

import (
	"fmt"
	"image/color"
)

type HalfBlockChar interface {
	IRC() string
}

type HalfBlockHexColor struct {
	HalfBlockChar
	Top    color.NRGBA
	Bottom color.NRGBA
}

// is this even correct? it renders strange
// better to use color codes i suppose

func colorHex(c color.NRGBA) string {
	return fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B)
}

func (p *HalfBlockHexColor) IRC() string {
	return fmt.Sprintf("%c%s,%s", 0x04, colorHex(p.Top), colorHex(p.Bottom))
}

type HalfBlockColorCode struct {
	HalfBlockChar
	Top    uint8
	Bottom uint8
}

func validColorCode(code uint8) bool {
	return code >= 16 && code <= 98
}

func (p *HalfBlockColorCode) IRC() string {
	if p.Top == 0 && p.Bottom == 0 {
		return fmt.Sprintf("%c ", 0x0f)
	}

	if p.Top > 0 && p.Bottom == 0 {
		// only top
		if !validColorCode(p.Top) {
			panic(fmt.Errorf("invalid top color code: %d", p.Top))
		}
		return fmt.Sprintf("%c%c%02d▀", 0x0f, 0x03, p.Top)
	} else if p.Bottom > 0 && p.Top == 0 {
		// only bottom
		if !validColorCode(p.Bottom) {
			panic(fmt.Errorf("invalid bottom color code: %d", p.Bottom))
		}
		return fmt.Sprintf("%c%c%02d▄", 0x0f, 0x03, p.Bottom)
	}

	if !validColorCode(p.Top) {
		panic(fmt.Errorf("invalid top color code: %d", p.Top))
	}

	if !validColorCode(p.Bottom) {
		panic(fmt.Errorf("invalid bottom color code: %d", p.Bottom))
	}

	return fmt.Sprintf("%c%02d,%02d▀", 0x03, p.Top, p.Bottom)
}

type HalfBlockImage [][]HalfBlockChar

func (img HalfBlockImage) IRC() string {
	out := ".\n" // or it'll display weird on weechat android
	for _, row := range img {
		for x := range row {
			if row[x] != nil {
				out += row[x].IRC()
			}
		}
		out += "\n"
	}
	return out
}

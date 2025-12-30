package ircf

import (
	"fmt"
)

// https://modern.ircdocs.horse/formatting

const (
	codeBold = 0x02
	// codeItalic        = 0x1d
	// codeUnderline     = 0x1f
	// codeStrikethrough = 0x1e
	// codeMonospace     = 0x11
	codeColor = 0x03
	// codeHexColor      = 0x04
	// codeReverseColor  = 0x16
	codeReset = 0x0f
)

type Format struct {
	bold bool
	fg   uint8
	bg   uint8
}

// we want to make copies as we're currying and might want to mutate

func (f Format) Bold() Format {
	f.bold = true
	return f
}

func Bold() Format {
	return Format{}.Bold()
}

func (f Format) Color(fg uint8, bg ...uint8) Format {
	f.fg = fg
	if len(bg) > 0 {
		f.bg = bg[0]
	}
	return f
}

func Color(fg uint8, bg ...uint8) Format {
	return Format{}.Color(fg, bg...)
}

// outputting

func (f Format) Prefix() string {
	out := ""
	if f.bold {
		out += string(byte(codeBold))
	}
	if f.fg > 0 {
		out += fmt.Sprintf("%c%d", codeColor, f.fg)
		if f.bg > 0 {
			out += fmt.Sprintf(",%d", f.bg)
		}
	}
	return out
}

func (f Format) Format(msg string) string {
	return fmt.Sprintf("%s%s%c", f.Prefix(), msg, codeReset)
}

// predefined

// my client prints bold as purple, so force it to white
// too bad for light mode users i suppose
var BoldWhite = Bold().Color(98)

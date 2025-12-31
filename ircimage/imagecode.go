package ircimage

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"

	"github.com/disintegration/imaging"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/sync/semaphore"
)

var (
	//go:embed bluenoise.png
	blueNoisePNG   []byte
	blueNoiseImage image.Image

	// https://modern.ircdocs.horse/formatting
	// add 16 to the index. ends at 98. total colors: 83

	paletteHex = []uint32{
		0x470000, 0x472100, 0x474700, 0x324700,
		0x004700, 0x00472c, 0x004747, 0x002747,
		0x000047, 0x2e0047, 0x470047, 0x47002a,
		0x740000, 0x743a00, 0x747400, 0x517400,
		0x007400, 0x007449, 0x007474, 0x004074,
		0x000074, 0x4b0074, 0x740074, 0x740045,
		0xb50000, 0xb56300, 0xb5b500, 0x7db500,
		0x00b500, 0x00b571, 0x00b5b5, 0x0063b5,
		0x0000b5, 0x7500b5, 0xb500b5, 0xb5006b,
		0xff0000, 0xff8c00, 0xffff00, 0xb2ff00,
		0x00ff00, 0x00ffa0, 0x00ffff, 0x008cff,
		0x0000ff, 0xa500ff, 0xff00ff, 0xff0098,
		0xff5959, 0xffb459, 0xffff71, 0xcfff60,
		0x6fff6f, 0x65ffc9, 0x6dffff, 0x59b4ff,
		0x5959ff, 0xc459ff, 0xff66ff, 0xff59bc,
		0xff9c9c, 0xffd39c, 0xffff9c, 0xe2ff9c,
		0x9cff9c, 0x9cffdb, 0x9cffff, 0x9cd3ff,
		0x9c9cff, 0xdc9cff, 0xff9cff, 0xff94d3,
		0x000000, 0x131313, 0x282828, 0x363636,
		0x4d4d4d, 0x656565, 0x818181, 0x9f9f9f,
		0xbcbcbc, 0xe2e2e2, 0xffffff,
	}
	palette []mgl32.Vec3
)

func init() {
	var err error
	blueNoiseImage, err = imaging.Decode(bytes.NewReader(blueNoisePNG))
	if err != nil {
		panic("failed to decode blue noise: " + err.Error())
	}

	palette = make([]mgl32.Vec3, len(paletteHex))
	for i := range paletteHex {
		palette[i][0] = float32(paletteHex[i]&0xff0000>>16) / 0xff
		palette[i][1] = float32(paletteHex[i]&0xff00>>8) / 0xff
		palette[i][2] = float32(paletteHex[i]&0xff) / 0xff
	}
}

func sampleThreshold(x, y int, n uint) float32 {
	x = GLSLModi(x, blueNoiseImage.Bounds().Dx())
	y = GLSLModi(y, blueNoiseImage.Bounds().Dy())
	r, g, b, _ := blueNoiseImage.At(x, y).RGBA()
	value := (float32(r) + float32(g) + float32(b)) / 3 / 0xffff
	return value * float32(n-1)
}

// https://www.shadertoy.com/view/dlcGzN

// https://chilliant.blogspot.com/2012/08/srgb-approximations-for-hlsl.html
func sRGBtoLinear(c mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		c.X() * (c.X()*(c.X()*0.305306011+0.682171111) + 0.012522878),
		c.Y() * (c.Y()*(c.Y()*0.305306011+0.682171111) + 0.012522878),
		c.Z() * (c.Z()*(c.Z()*0.305306011+0.682171111) + 0.012522878),
	}
}

func getLuminance(c mgl32.Vec3) float32 {
	return c.X()*0.299 + c.Y()*0.587 + c.Z()*0.114
}

func getClosestColorIndex(inputColor mgl32.Vec3) int {
	closestDistance := float32(math.MaxFloat32)
	closestColor := 0
	for i := range paletteHex {
		difference := inputColor.Sub(sRGBtoLinear(palette[i]))
		distance := difference.Dot(difference)
		if distance < closestDistance {
			closestDistance = distance
			closestColor = i
		}
	}
	return closestColor
}

func nrgbColorToFloat(c color.NRGBA) mgl32.Vec3 {
	return mgl32.Vec3{float32(c.R) / 0xff, float32(c.G) / 0xff, float32(c.B) / 0xff}
}

func floatColorToNrgb(c mgl32.Vec3) color.NRGBA {
	return color.NRGBA{
		R: uint8(c.X() * 0xff),
		G: uint8(c.Y() * 0xff),
		B: uint8(c.Z() * 0xff),
		A: 0xff,
	}
}

func optimizedKnoll(
	img *image.NRGBA, x, y int, n uint, errorFactor float32,
) (uint8, bool) {
	color := nrgbColorToFloat(img.NRGBAAt(x, y))

	// accumulate the frequencies for each palette colour
	frequency := make([]int, len(palette))
	quantError := mgl32.Vec3{0, 0, 0}
	colorLinear := sRGBtoLinear(color)

	for range n {
		goalColor := colorLinear.Add(quantError.Mul(errorFactor))
		closestColor := getClosestColorIndex(goalColor)
		frequency[closestColor] += 1
		quantError = quantError.Add(
			colorLinear.Sub(sRGBtoLinear(palette[closestColor])),
		)
	}

	// select the output colour by accumulating the frequencies
	// until the candidate is found
	randomValue := int(sampleThreshold(x, y, n))
	cumulativeSum := 0

	for i := range len(palette) {
		cumulativeSum += frequency[i]
		if randomValue < cumulativeSum {
			return uint8(i), true
		}
	}

	return 0, false
}

func unoptimizedKnoll(
	img *image.NRGBA, x, y int,
	n uint, errorFactor float32,
) (uint8, bool) {
	color := nrgbColorToFloat(img.NRGBAAt(x, y))

	// fill the candidate array
	candidates := make([]int, n)
	quantError := mgl32.Vec3{0, 0, 0}
	colorLinear := sRGBtoLinear(color)

	for i := range n {
		goalColor := colorLinear.Add(quantError.Mul(errorFactor))
		closestColor := getClosestColorIndex(goalColor)

		candidates[i] = closestColor
		quantError = quantError.Add(
			colorLinear.Sub(sRGBtoLinear(palette[closestColor])),
		)
	}

	// sort the candidate array by luminance (bubble sort)
	for i := n - 1; i > 0; i-- {
		for j := range i {
			if getLuminance(palette[candidates[j]]) >
				getLuminance(palette[candidates[j+1]]) {
				// swap candidates
				t := candidates[j]
				candidates[j] = candidates[j+1]
				candidates[j+1] = t
			}
		}
	}

	// select from the candidate array, using the value in the threshold matrix
	index := int(sampleThreshold(x, y, n))
	return uint8(candidates[index]), true
}

func thomasKnollDither(
	img *image.NRGBA,
	n uint, errorFactor float32,
) [][]uint8 {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	var out [][]uint8 = make([][]uint8, h)
	for y := range h {
		out[y] = make([]uint8, w)
	}

	writePixel := func(x, y int, i uint8) {
		// img.Set(x, y, floatColorToNrgb(colorCodes[i]))
		if int(i) > len(palette) {
			panic(fmt.Errorf("index out of color range: %d", i))
		}
		out[y][x] = i + 16
	}

	doPixel := func(x, y int) {
		i, ok := optimizedKnoll(img, x, y, n, errorFactor)
		if ok {
			writePixel(x, y, i)
			return
		}
		// i, ok := unoptimizedKnoll(img, x, y, n, errorFactor)
		// if ok {
		// 	writePixel(x, y, i)
		// 	return
		// }
	}

	ctx := context.Background()
	maxWorkers := int64(runtime.GOMAXPROCS(0))
	sem := semaphore.NewWeighted(maxWorkers)

	for y := range h {
		for x := range w {
			err := sem.Acquire(ctx, 1)
			if err != nil {
				panic(err)
			}
			go func() {
				defer sem.Release(1)
				doPixel(x, y)
			}()
		}
	}

	err := sem.Acquire(ctx, maxWorkers)
	if err != nil {
		panic(err)
	}

	return out
}

// func saveDitheredImage(img [][]uint8, w, h int, filename string) {
// 	rgbImg := image.NewNRGBA(image.Rectangle{Max: image.Pt(w, h)})
// 	for y := range h {
// 		for x := range w {
// 			hex := paletteHex[img[y][x]-16]
// 			rgbImg.SetNRGBA(x, y, color.NRGBA{
// 				R: uint8(hex >> 16 & 0xff),
// 				G: uint8(hex >> 8 & 0xff),
// 				B: uint8(hex & 0xff),
// 				A: 0xff,
// 			})
// 		}
// 	}
// 	err := imaging.Save(rgbImg, filename)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// n: number of iterations per fragment (default: 32, higher == more samples)
// errorFactor: quantization error coefficient (default: 0.8, 0 == no dithering)
func ConvertImageWithColorCodesDither(
	img *image.NRGBA, n uint, errorFactor float32,
) (out HalfBlockImage, err error) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	if h%2 != 0 {
		return nil, errors.New("height must be a multiple of 2")
	}

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		switch v := r.(type) {
		case string:
			err = errors.New(v)
		case error:
			err = v
		default:
			err = errors.New("unknown panic")
		}
	}()

	dithered := thomasKnollDither(img, n, errorFactor)
	// saveDitheredImage(dithered, w, h, "test.png")

	out = make([][]HalfBlockChar, h/2)

	for y := 0; y < h; y += 2 {
		out[y/2] = make([]HalfBlockChar, w)
		for x := range w {
			out[y/2][x] = &HalfBlockColorCode{
				Top: dithered[y][x], Bottom: dithered[y+1][x],
			}
		}
	}

	return
}

func ConvertImageWithColorCodesNodither(
	img *image.NRGBA,
) (out HalfBlockImage, err error) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	if h%2 != 0 {
		return nil, errors.New("height must be a multiple of 2")
	}

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		switch v := r.(type) {
		case string:
			err = errors.New(v)
		case error:
			err = v
		default:
			err = errors.New("unknown panic")
		}
	}()

	var pixels [][]uint8 = make([][]uint8, h)
	for y := range h {
		pixels[y] = make([]uint8, w)
	}

	for y := range h {
		for x := range w {
			color := nrgbColorToFloat(img.NRGBAAt(x, y))
			color = sRGBtoLinear(color)
			pixels[y][x] = 16 + uint8(getClosestColorIndex(color))
		}
	}

	out = make([][]HalfBlockChar, h/2)

	for y := 0; y < h; y += 2 {
		out[y/2] = make([]HalfBlockChar, w)
		for x := range w {
			out[y/2][x] = &HalfBlockColorCode{
				Top: pixels[y][x], Bottom: pixels[y+1][x],
			}
		}
	}

	return
}

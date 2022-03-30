package main

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
	"math"
	"os"

	"github.com/fogleman/gg"
	"github.com/pbberlin/dbg"
)

/* var palette = []color.Color{
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0x00, 0x00, 0xff, 0xff},
	color.RGBA{0x00, 0xff, 0x00, 0xff},
	color.RGBA{0x00, 0xff, 0xff, 0xff},
	color.RGBA{0xff, 0x00, 0x00, 0xff},
	color.RGBA{0xff, 0x00, 0xff, 0xff},
	color.RGBA{0xff, 0xff, 0x00, 0xff},
	color.RGBA{0xff, 0xff, 0xff, 0xff},
}
*/
func makePalette(sz int) []color.Color {

	bs := uint8(0xff)
	pal := make([]color.Color, 0, sz)

	rgb := [3]uint8{}

	alpha := bs
	//   8 / 8 =>  1
	// 256 / 8 => 32
	numShades := float64(sz / 8)

	for ci := 0; ci < sz; ci++ { // color index - ci

		shade := math.Floor(float64(ci) / numShades) // 0...7

		// log.Printf("color #%v", ci)
		for cci := 0; cci < 3; cci++ { // color component index - ci
			v := byte(ci)
			if (v>>cci)&1 == 1 {
				rgb[cci] = uint8(0xff) - uint8(shade*numShades)

			} else {
				rgb[cci] = uint8(0x00)
			}
			if ci%2 == 0 && (ci < 8 || ci > 251 || ci%40 == 0) {
				// log.Printf("color %3v:  bit %v to %3v", ci, cci, rgb[cci])
			}
		}

		col := color.RGBA{rgb[0], rgb[1], rgb[2], alpha}
		pal = append(pal, col)
	}

	_ = dbg.Dump2String(pal)

	return pal

}

func renderIntoPalettedImage(c *gg.Context) *image.Paletted {

	src := c.Image()
	var dst *image.Paletted
	dst = image.NewPaletted(src.Bounds(), makePalette(256))
	// dst = image.NewPaletted(src.Bounds(), palette.WebSafe)
	dst = image.NewPaletted(src.Bounds(), palette.Plan9)

	if false {
		// Floyd-Steinberg makes artifacts but Image with good colors and doesn't have gradient issue. Only issue with dots.
		drawer := draw.FloydSteinberg
		drawer.Draw(dst, dst.Bounds(), src, image.Point{0, 0})
	} else {
		// file size seven times smaller than FloydSteinberg
		// little quality loss
		draw.Draw(dst, dst.Bounds(), src, image.Point{0, 0}, draw.Over)
	}
	return dst

}

type Circle struct {
	X, Y, R float64
}

func (c *Circle) Brightness(x, y float64) uint8 {
	var dx, dy float64 = c.X - x, c.Y - y
	d := math.Sqrt(dx*dx+dy*dy) / c.R
	if d > 1 {
		return 0
	} else {
		return 128
		// return 255
	}
}

func saveAnimGIF(images []*image.Paletted, delays []int, fn string) {
	f, err := os.OpenFile(
		fmt.Sprintf("./out/%v.gif", fn),
		os.O_WRONLY|os.O_CREATE,
		0777,
	)
	if err != nil {
		log.Printf("cannot open file for gif: %v", err)
		return
	}
	defer f.Close()
	err = gif.EncodeAll(f,
		&gif.GIF{
			Image: images,
			Delay: delays, // 100ths of a second
		},
	)
	if err != nil {
		log.Printf("cannot encode images into anim gif container: %v", err)
		return
	}
}

func animGIF1() {

	var w, h int = 240, 240
	var hw, hh float64 = float64(w / 2), float64(h / 2)

	pal := makePalette(8)
	var images []*image.Paletted
	var delays []int

	circles := []*Circle{{}, {}, {}}
	steps := 20

	for step := 0; step < steps; step++ {

		img := image.NewPaletted(image.Rect(0, 0, w, h), pal)
		images = append(images, img)
		delays = append(delays, 0)

		θ := 2.0 * math.Pi / float64(steps) * float64(step)
		for i, circle := range circles {
			θ0 := 2 * math.Pi / 3 * float64(i)
			circle.X = hw - 40*math.Sin(θ0) - 20*math.Sin(θ0+θ)
			circle.Y = hh - 40*math.Cos(θ0) - 20*math.Cos(θ0+θ)
			circle.R = 50
		}

		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				img.Set(x, y, color.RGBA{
					circles[0].Brightness(float64(x), float64(y)),
					circles[1].Brightness(float64(x), float64(y)),
					circles[2].Brightness(float64(x), float64(y)),
					255,
				})
			}
		}
	}

	saveAnimGIF(images, delays, "anim-1.gif")

}

func animGIF2() {

	var w, h int = 240, 240
	var hw, hh float64 = float64(w / 2), float64(h / 2)

	pal := makePalette(256)
	var images []*image.Paletted
	var delays []int

	circles := []*Circle{{}, {}, {}}
	steps := 20

	//
	for step := 0; step < steps; step++ {

		// log.Printf("step %2v of %v", step, steps)

		img := image.NewPaletted(image.Rect(0, 0, w, h), pal)
		images = append(images, img)

		if step < steps/2 {
			delays = append(delays, step)
		} else {
			delays = append(delays, steps-step)
		}

		θ := 2.0 * math.Pi / float64(steps) * float64(step)
		for i, circle := range circles {
			θ0 := 2 * math.Pi / 3 * float64(i)
			circle.X = hw - 40*math.Sin(θ0) - 20*math.Sin(θ0+θ)
			circle.Y = hh - 40*math.Cos(θ0) - 20*math.Cos(θ0+θ)
			circle.R = 50
		}

		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				img.Set(x, y, color.RGBA{
					circles[0].Brightness(float64(x), float64(y)),
					circles[1].Brightness(float64(x), float64(y)),
					circles[2].Brightness(float64(x), float64(y)),
					255,
				})
			}
		}
	}

	saveAnimGIF(images, delays, "anim-2.gif")

}

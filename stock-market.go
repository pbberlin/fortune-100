package main

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"log"

	"github.com/fogleman/gg"
)

var c *gg.Context
var w, h float64

func init() {
	w = 1024
	h = 768
	c = gg.NewContext(int(w), int(h))
}

func ci(x, y, rad int) {
	c.DrawCircle(float64(x)/100*w, float64(y)/100*h, float64(rad)/100*w)
}

func stockMarket1() {

	ci(50, 50, 20)
	// c.SetRGB(0.1, 0.5, 0)
	c.SetRGB(0.1, 0.5, 0)
	c.Fill()

	ci(10, 10, 12)
	c.SetRGB(0.5, 0.1, 0)
	c.Fill()

	err := c.SavePNG("./out/stock-market-1.png")
	if err != nil {
		log.Printf("SavePNG reports %v", err)
	}

}

func stockMarket2Init(w, h float64) *gg.Context {
	var c *gg.Context
	c = gg.NewContext(int(w), int(h))
	return c
}

func stockMarket2() {

	pal := makePalette(256)
	var images []*image.Paletted
	var delays []int

	var w, h float64
	w = 1024
	h = 768
	steps := 4

	var ctx *gg.Context
	drawC := func(x, y, rad int) {
		// log.Printf("circle drawn")
		ctx.DrawCircle(float64(x)/100*w, float64(y)/100*h, float64(rad)/100*w)
	}

	//
	for step := 0; step < steps; step++ {

		// log.Printf("step %2v of %v", step, steps)

		ctx = stockMarket2Init(w, h)

		factor := 5 * step

		drawC(50+factor, 50+factor, 20+factor)
		ctx.SetRGB(0.1, 0.5, 0)
		ctx.Fill()

		drawC(10+factor, 10+factor, 12+factor)
		ctx.SetRGB(0.5, 0.1, 0)
		ctx.Fill()

		if true {
			fn := fmt.Sprintf("./out/sm2_%02v.png", step)
			ctx.SavePNG(fn)
		}

		if true {
			src := ctx.Image()

			var dst *image.Paletted
			// dst = image.NewPaletted(image.Rect(0, 0, int(w), int(h)), palette.Plan9)
			dst = image.NewPaletted(src.Bounds(), palette.WebSafe)
			dst = image.NewPaletted(src.Bounds(), pal)
			dst = image.NewPaletted(src.Bounds(), palette.Plan9)

			// image.Point{0, 0} instead of image.ZP

			if false {
				// Floyd-Steinberg makes artifacts but Image with good colors and doesn't have gradient issue. Only issue with dots.
				drawer := draw.FloydSteinberg
				drawer.Draw(dst, dst.Bounds(), src, image.Point{0, 0})
			} else {
				// file size seven times smaller than FloydSteinberg
				// little quality loss
				draw.Draw(dst, dst.Bounds(), src, image.Point{0, 0}, draw.Over)
			}

			images = append(images, dst)

			if step < steps/2 {
				factor := 10
				delays = append(delays, step*factor)
			} else {
				delays = append(delays, (steps-step)*factor)
			}
		}

	}

	saveAnimGIF(images, delays, "stock-market-2")

}

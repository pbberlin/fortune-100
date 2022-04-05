package main

import (
	"fmt"
	"image"
	"log"

	"github.com/fogleman/gg"
)

func stockMarket2(rksYears RankingsYears, companiesByName map[string]company) {

	var images []*image.Paletted
	var delays []int

	var w, h float64
	w = 1024
	h = 768

	years := rksYears.Years()

	c := gg.NewContext(int(w), int(h))
	fontSize := 96.0
	fontSize = 14.0
	fontSize = 12.0
	if err := c.LoadFontFace("./out/arial.ttf", fontSize); err != nil {
		log.Fatalf("Cannot load font: %v", err)
	}
	scale100 := float64(w) / float64(100)

	// funcs as closures to reduce number of parameters
	cntr2 := 0
	draw := func(x, y, rad, revenue, maxRev float64) {

		rd := revenue / maxRev * rad // max rad

		rdScaled := 6 * rd
		if rdScaled > rad {
			rdScaled = rad
		}

		cntr2++
		if cntr2%13 == 0 {
			log.Printf("%11v %11v - %5.2v - %5.2v - %5.2v", revenue, maxRev, revenue/maxRev, revenue/maxRev*rad, rdScaled)
		}

		c.DrawCircle(scale100*x, scale100*y, scale100*rdScaled)
	}
	text := func(x, y, w float64, s string) {
		c.SetRGB(0.95, 0.95, 0.95)
		// c.DrawString(s, scale100*x, scale100*y+c.FontHeight())
		// c.DrawStringAnchored(s, scale100*x, scale100*y, 0.5, 0.5)
		c.DrawStringWrapped(
			s,
			scale100*x, scale100*y,
			0.5, 0.5,
			w,
			1.3,
			gg.AlignCenter,
		)
	}

	//
	//
	cntr := -1
	for _, yr := range years {

		cntr++

		c.SetRGB(0.1, 0.1, 0.1)
		c.Clear()

		c.SetRGB(0.8, 0.8, 0.8)
		c.DrawString(fmt.Sprintf("Yr %v", yr), 5, 5+c.FontHeight())

		cx := 0.0
		bx := 16.0 // box size - displacement
		bx = 12.0
		bx = 9.0

		row := 0.0
		for i := 0; i < len(rksYears[cntr].Rankings); i++ {

			rv := rksYears[cntr].Rankings[i].Revenue
			nm := rksYears[cntr].Rankings[i].Name

			cx += bx

			if cx+bx >= 100 {
				cx = 0
				row++
				// log.Printf("row %3v", row)
			}

			x := cx + bx/2

			// log.Printf("  cx is %3v - x %3v", cx, x)

			y := 100 - bx - (row * bx)
			// c.DrawCircle(x, y, 8.0)
			draw(x, y, bx/2, rv, rksYears[cntr].Max)
			// log.Printf("drawing %4v %4v - %v", x, y, cl)
			c.SetColor(companiesByName[nm].Color)
			c.Fill()

			text(x, y, bx, nm)

		}

		// save to PNG
		if true {
			fn := fmt.Sprintf("./out/sm2_%02v.png", yr)
			c.SavePNG(fn)
		}

		images = append(images, renderIntoPalettedImage(c))

		elongation := 50
		if cntr < len(years)/2 {
			delays = append(delays, cntr*elongation)
		} else {
			delays = append(delays, (len(years)-cntr)*elongation)
		}

	}

	saveAnimGIF(images, delays, "stock-market-2")

}

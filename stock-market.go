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
	// if err := c.LoadFontFace("./out/arialbd.ttf", 96); err != nil {
	if err := c.LoadFontFace("./out/arialbd.ttf", 14); err != nil {
		log.Fatalf("Cannot load font: %v", err)
	}
	scale100 := float64(w) / float64(100)
	draw := func(x, y, r float64) {
		c.DrawCircle(scale100*x, scale100*y, scale100*r)
	}
	text := func(x, y float64, s string) {
		c.SetRGB(0.8, 0.8, 0.8)
		c.DrawString(
			s,
			scale100*x,
			scale100*y+c.FontHeight(),
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

		row := 0.0
		cx := 0.0
		dp := 16.0 // displacement
		for i := 0; i < len(rksYears[cntr].Rankings); i++ {

			rv := rksYears[cntr].Rankings[i].Revenue
			nm := rksYears[cntr].Rankings[i].Name

			_, _ = rv, nm

			if cx+dp > 100 {
				cx = 0
				row++
			} else {
				cx += dp
			}

			x := cx + dp/2

			y := 100 - dp - (row * dp)
			cl := companiesByName[nm].Color
			// c.DrawCircle(x, y, 8.0)
			draw(x, y, dp/2*0.8)
			// log.Printf("drawing %4v %4v - %v", x, y, cl)
			c.SetColor(cl)
			c.Fill()

			text(x, y, companiesByName[nm].Name)

		}

		// save to PNG
		if true {
			fn := fmt.Sprintf("./out/sm2_%02v.png", yr)
			c.SavePNG(fn)
		}

		images = append(images, renderIntoPalettedImage(c))

		elongation := 10
		if yr < len(years)/2 {
			delays = append(delays, yr*elongation)
		} else {
			delays = append(delays, (len(years)-yr)*elongation)
		}

	}

	saveAnimGIF(images, delays, "stock-market-2")

}

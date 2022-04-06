package main

import (
	"fmt"
	"image"
	"image/color"
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
	fontSize = 12.0
	if err := c.LoadFontFace("./out/arial.ttf", fontSize); err != nil {
		log.Fatalf("Cannot load font: %v", err)
	}

	scale100 := float64(w) / float64(100)
	if h < w {
		scale100 = float64(h) / float64(100) // shorter side dominates
	}

	// funcs as closures to reduce number of parameters
	drwC := func(x, y, bx, revenue, maxRev float64) {

		boxRad := bx / 2
		rd := revenue / maxRev * boxRad // max rad

		scaleUp := 4.0
		scaleUp = 2.0
		scaleUp = 1.0
		rdScaled := scaleUp * rd
		if rdScaled > boxRad {
			// rdScaled = boxRad
		}

		// c.DrawCircle(scale100*x, scale100*y, scale100*rdScaled)
		c.DrawCircle(
			scale100*x,
			scale100*(y+boxRad),
			scale100*rdScaled,
		)
	}
	drwTxt := func(x, y, bx float64, s string) {
		c.SetRGB(0.95, 0.95, 0.95)
		// c.DrawString(s, scale100*x, scale100*y+c.FontHeight())
		// c.DrawStringAnchored(s, scale100*x, scale100*y, 0.5, 0.5)
		c.DrawStringWrapped(
			s,
			scale100*x, scale100*(y+bx),
			// 0.5, 0.5,
			0.5, 0.99,
			bx*0.95,
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
		bx = 9.0 // eleven per row
		bx = 9.8 // ten per row - slightly

		row := 0.0
		// log.Print(" ")
		// log.Printf("row %v", row)

		for i := 0; i < len(rksYears[cntr].Rankings); i++ {

			// ten per row
			if len(rksYears[cntr].Rankings)-i > 100 {
				// equalize number of rankings between 101 and 100
				continue
			}

			// eleven per row
			// if len(rksYears[cntr].Rankings)-i > 99 {
			// 	// equalize number of rankings between 101 and 100
			// 	continue
			// }

			rv := rksYears[cntr].Rankings[i].Revenue
			nm := rksYears[cntr].Rankings[i].Name
			sh := rksYears[cntr].Rankings[i].Short

			if cx+bx >= 100 {
				cx = 0
				row++
				// log.Printf("row %v", row)
			}

			cx += bx

			x := cx + bx/2

			y := 100 - bx - (row * bx)

			// if i%5 == 0 {
			// 	log.Printf("  cx %3.0f    x %3.0f    row %2.0v    y %3.0f", cx, x, row, y)
			// }

			if false {
				c.DrawRectangle(scale100*(x-bx/2+1), scale100*(y+1), scale100*(bx-2), scale100*(bx-2))
				c.SetColor(color.RGBA{44, 44, 44, 55})
				c.Fill()
			}

			// c.DrawCircle(x, y, 8.0)
			// drwC(x, y, bx, rv, rksYears[cntr].MaxTotal)
			drwC(x, y, bx, rv, rksYears[cntr].Quant95Total)
			// log.Printf("drawing %4v %4v - %v", x, y, cl)
			c.SetColor(companiesByName[nm].Color)
			c.Fill()

			drwTxt(x, y, bx, sh)

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

		// break

	}

	saveAnimGIF(images, delays, "stock-market-2")

}

package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/fogleman/gg"
)

func stockMarket2() {

	rksyrs := readRankingsYears()

	rksyrs.QuantilesTotal()

	companiesByName := readCompaniesByYears()

	var images []*image.Paletted
	var delays []int

	var w, h float64
	w = 1024
	h = 768

	// all rendering arguments are standardized to
	//   100 units of canvas height;
	//   thus, 133.3 is the according max width
	wOverH := 1024 * 100.0 / 768

	years := rksyrs.Years()

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
	drwC := func(x, y, boxRad, revenue, mxRv float64) {

		cRad := revenue / mxRv * boxRad // circle rad

		c.DrawCircle(
			scale100*x,
			scale100*(y+boxRad),
			scale100*cRad,
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

	newRowOnQuant := false

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
		cy := 0.0
		bx := 7.8 // 133 / 7.8 => roughly 17

		lpQuantile := 0
		maxRev := rksyrs.Qs[90]

		// log.Print(" ")

		for i := 0; i < len(rksyrs.RkgsYear[cntr].Rankings); i++ {

			// ten per row
			if len(rksyrs.RkgsYear[cntr].Rankings)-i > 100 {
				// equalize number of rankings between 101 and 100
				continue
			}

			rv := rksyrs.RkgsYear[cntr].Rankings[i].Revenue
			nm := rksyrs.RkgsYear[cntr].Rankings[i].Name
			sh := rksyrs.RkgsYear[cntr].Rankings[i].Short

			rowFull := cx+bx >= wOverH
			newQuantile := rv > rksyrs.Qs[90] && lpQuantile != 90

			if rowFull {
				if !newRowOnQuant {
					cx = 0
					cy += bx // before computing new bx
				}
			}
			if rowFull || newQuantile {
				if newRowOnQuant {
					cx = 0
					cy += bx // before computing new bx

				}
			}
			if newQuantile {

				sizeUp := rksyrs.Qs[98] / rksyrs.Qs[90] * 0.98

				log.Printf("sizing up the box from %5.1f to %5.1f", bx, bx*sizeUp*1.2)

				bx *= sizeUp

				lpQuantile = 90
				maxRev = rksyrs.Qs[98]

				log.Printf("Yr %v-Quant chg %v", yr, i)
			}

			x := cx + bx/2
			cx += bx

			y := 100 - bx - cy

			// if i%5 == 0 {
			// 	log.Printf("  cx %3.0f    x %3.0f    row %2.0v    y %3.0f", cx, x, row, y)
			// }

			if true {
				c.DrawRectangle(scale100*(x-bx/2+0.3), scale100*(y+0.3), scale100*(bx-0.6), scale100*(bx-0.6))
				c.SetColor(color.RGBA{32, 32, 32, 80})
				c.Fill()
			}

			drwC(x, y, bx/2, rv, maxRev)
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

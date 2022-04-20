package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/fogleman/gg"
)

func loadFont(c *gg.Context, fontSize float64) {
	// fontSize := 96.0
	// fontSize = 12.0
	if err := c.LoadFontFace("./out/arial.ttf", fontSize); err != nil {
		log.Fatalf("Cannot load font: %v", err)
	}

}

func stockMarket2() {

	rksyrs := readRankingsYears()

	rksyrs.QuantilesTotal()

	companiesByName := readCompaniesByYears()

	var images []*image.Paletted
	var delays []int

	var w, h float64
	w = 1024
	h = 768

	// initial box size,
	// 133 / 7.8   => roughly 17
	bxBase := 7.8
	bxBase = 7.35 // 18
	bxBase = 6.63 // 20
	bxBase = 6.04 // 22
	bxBase = 5.54 // 24
	bxBaseRad := bxBase / 2

	baseQuant := rksyrs.Qs[1]

	// all rendering arguments are standardized to
	//   100 units of canvas height;
	//   thus, 133.3 is the according max width
	wOverH := 1024 * 100.0 / 768 // width over height

	years := rksyrs.Years()

	c := gg.NewContext(int(w), int(h))
	loadFont(c, 12)

	scale100 := float64(w) / float64(100)
	if h < w {
		scale100 = float64(h) / float64(100) // shorter side dominates
	}

	// funcs as closures to reduce number of parameters
	drwC := func(
		x, y, boxRad float64,
		// bxSizeUp float64,
		// quantRevPrv float64,
		// quantRev float64,
		companyRev float64,
	) {

		zeroToOne := math.Sqrt(companyRev / baseQuant.Rev)
		cRad := zeroToOne * bxBaseRad // circle radius
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

	//
	//
	contRows := true // continuous rows - even on new quantile
	// contRows = false
	frameCntr := -1
	for _, yr := range years {
		frameCntr++

		// empty existing conext
		c.SetRGB(0.001, 0.001, 0.001)
		c.Clear()

		cx := 0.0 // distance from left
		cy := 0.0 // distance from bottom

		bx := bxBase

		lastBox := bx

		bxMrg := 0.1

		// log.Print(" ")
		quant := baseQuant // current quantile

		for i := 0; i < len(rksyrs.RkgsYear[frameCntr].Rankings); i++ {

			if len(rksyrs.RkgsYear[frameCntr].Rankings)-i > 100 {
				// equalize number of rankings between 101 and 100;
				// if we start for instance with thirteen per row;
				// then the biggest circle has the same rank (100, not 101/100/102)
				continue
			}

			rv := rksyrs.RkgsYear[frameCntr].Rankings[i].Revenue
			nm := rksyrs.RkgsYear[frameCntr].Rankings[i].Name
			sh := rksyrs.RkgsYear[frameCntr].Rankings[i].Short

			newQuantile := rv > quant.Rev

			if newQuantile {
				// tentative
				newQuant := rksyrs.Qs.Next(quant.Q)
				bxSizeUp := math.Sqrt(newQuant.Rev / quant.Rev)
				if frameCntr == 0 {
					log.Printf("yr %v - quant%03v to %3v: box sizing from %5.1f to %5.1f", yr, quant.Q, newQuant.Q, bx, bx*bxSizeUp)
				}
				lastBox = bx
				bx *= bxSizeUp
				quant = newQuant
			}

			rowOverflow := cx+bx >= wOverH

			if contRows {
				if rowOverflow && newQuantile {
					cx = 0
					cy += lastBox // before computing new bx
				}
				if rowOverflow && !newQuantile {
					cx = 0
					cy += bx
				}
			} else {
				if rowOverflow || newQuantile {
					if newQuantile {
						cx = 0
						cy += lastBox // before computing new bx
					}
					if !newQuantile {
						cx = 0
						cy += bx
					}
				}
			}

			x := cx + bx/2
			cx += bx
			y := 100 - bx - cy

			// if i%5 == 0 {
			// 	log.Printf("  cx %3.0f    x %3.0f    row %2.0v    y %3.0f", cx, x, row, y)
			// }

			if true {
				c.DrawRectangle(scale100*(x-bx/2+bxMrg), scale100*(y+bxMrg), scale100*(bx-2*bxMrg), scale100*(bx-2*bxMrg))
				c.SetColor(color.RGBA{32, 32, 32, 80})
				c.Fill()
			}

			drwC(
				x, y, bx/2,
				rv,
			)
			// log.Printf("drawing %4v %4v - %v", x, y, cl)
			c.SetColor(companiesByName[nm].Color)
			c.Fill()

			drwTxt(x, y, bx, sh)

		}
		//
		// frame label
		loadFont(c, 32)
		c.SetRGB(0.8, 0.8, 0.8)
		c.DrawString(fmt.Sprintf("#%v - year %v", frameCntr+1, yr), 5, 5+c.FontHeight())
		loadFont(c, 12) // reset font

		//
		// save to PNG
		if true {
			fn := fmt.Sprintf("./out/sm2_%02v.png", yr)
			c.SavePNG(fn)
		}

		images = append(images, renderIntoPalettedImage(c))

		elongation := 50
		if frameCntr < len(years)/2 {
			delays = append(delays, frameCntr*elongation)
		} else {
			delays = append(delays, (len(years)-frameCntr)*elongation)
		}

		// break

	}

	saveAnimGIF(images, delays, "stock-market-2")

}

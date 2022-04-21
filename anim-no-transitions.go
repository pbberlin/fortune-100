package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/fogleman/gg"
)

func animationsNoTransition() {

	// fetching data
	rksyrs := readRankingsYears()
	rksyrs.QuantilesTotal()
	companiesByName := readCompaniesByYears()
	baseQuant := rksyrs.Qs[1] // quantile
	years := rksyrs.Years()

	//
	// output structures
	var images []*image.Paletted
	var delays []int

	// go graphics
	// 		we use a single context, cleaning it between frames
	c := gg.NewContext(int(w), int(h))
	loadFont(c, 12)

	sf := computeSF()

	// funcs as closures
	// having access to all global parameters above
	// reducing the number of parameters;
	// draw circle:
	drwC := func(
		x, y, boxRad float64,
		companyRev float64,
	) {
		zeroToOne := math.Sqrt(companyRev / baseQuant.Rev)
		cRad := zeroToOne * bxBaseRad // circle radius
		c.DrawCircle(
			sf*x,
			sf*(y+boxRad),
			sf*cRad,
		)

	}
	// drwTxt draws text centered at x, vertically bottomed, with a width of bx
	drwTxt := func(x, y, bx float64, s string) {
		c.SetRGB(0.95, 0.95, 0.95)
		// c.DrawString        (s, scale100*x, scale100*y+c.FontHeight())
		// c.DrawStringAnchored(s, scale100*x, scale100*y, 0.5, 0.5)
		c.DrawStringWrapped(
			s,
			sf*x, sf*(y+bx),
			// 0.5, 0.5,
			0.5, 0.99,
			bx*0.95,
			1.3,
			gg.AlignCenter,
		)
	}

	//
	//
	continuousRows := true // continuous rows - dont start new now on new quantile
	frameCntr := -1        // number of images
	for _, yr := range years {
		frameCntr++

		// empty existing context from previous frame drawings
		c.SetRGB(0.001, 0.001, 0.001)
		c.Clear()

		cx := 0.0 // distance from left
		cy := 0.0 // distance from bottom

		bx := bxBase
		lastBox := bx

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

			quantileOverflow := rv > quant.Rev
			if quantileOverflow {
				newQuant := quant
				for i := 0; ; i++ {
					newQuant = rksyrs.Qs.Next(newQuant.Q)
					if i > 0 {
						log.Printf("yr %v - jumping %v quantiles: comp %5v, rank %3v RV %5.0f - quant rev %5.0f",
							yr, i+1, sh, rksyrs.RkgsYear[frameCntr].Rankings[i].Rank, rv, newQuant.Rev,
						)
					}
					// quantile overflow - end
					if rv < newQuant.Rev || newQuant.Q == 100 {
						break
					}
				}
				bxSizeUp := math.Sqrt(newQuant.Rev / quant.Rev)
				if frameCntr == 0 || frameCntr == len(years)-1 {
					log.Printf("yr %v - quant%03v to %3v: box sizing from %5.1f to %5.1f", yr, quant.Q, newQuant.Q, bx, bx*bxSizeUp)
				}
				lastBox = bx
				bx *= bxSizeUp
				quant = newQuant
			}

			rowOverflow := cx+bx >= wOverH

			if continuousRows {
				if rowOverflow && quantileOverflow {
					cx = 0
					cy += lastBox // before computing new bx
				}
				if rowOverflow && !quantileOverflow {
					cx = 0
					cy += bx
				}
			} else {
				if rowOverflow || quantileOverflow {
					if quantileOverflow {
						cx = 0
						cy += lastBox // before computing new bx
					}
					if !quantileOverflow {
						cx = 0
						cy += bx
					}
				}
			}

			x := cx + bx/2
			cx += bx
			y := 100 - bx - cy

			// 	log.Printf("  cx %3.0f    x %3.0f    row %2.0v    y %3.0f", cx, x, row, y)

			// drawing a pale box - to make the box packing easy to spot
			if true {
				bxMrg := 0.1
				c.DrawRectangle(sf*(x-bx/2+bxMrg), sf*(y+bxMrg), sf*(bx-2*bxMrg), sf*(bx-2*bxMrg))
				c.SetColor(color.RGBA{32, 32, 32, 80})
				c.Fill()
			}

			drwC(
				x, y, bx/2,
				rv,
			)

			// painting the company name
			c.SetColor(companiesByName[nm].Color)
			c.Fill()
			drwTxt(x, y, bx, sh)

		}

		//
		// left top frame label
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

		// save to animated GIF structure
		images = append(images, renderIntoPalettedImage(c))
		elongation := 50
		if frameCntr < len(years)/2 {
			delays = append(delays, frameCntr*elongation)
		} else {
			delays = append(delays, (len(years)-frameCntr)*elongation)
		}

	}

	saveAnimGIF(images, delays, "stock-market-2")

}

package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/fogleman/gg"
)

func animationsTransitionStage3() {

	mainFrames := Load("mainFrames2.json")

	//
	// output structures
	var images []*image.Paletted
	var delays []int
	fontCol := color.RGBA{255, 255, 255, 255}

	// go graphics
	// 		we use a single context, cleaning it between frames
	c := gg.NewContext(int(w), int(h))
	loadFont(c, 12)

	sf := computeSF()

	drwC := func(
		x, y, rd float64,
		// companyRev float64,
		col color.RGBA,
		sh, longNammme string,
		bx float64,
		paling float64,
	) {

		// drawing a pale box - to make the box packing easy to spot
		if false {
			bxMrg := 0.2
			bxHalfMrg := bx/2 - bxMrg
			c.DrawRectangle(
				sf*(x-bxHalfMrg),
				sf*(y-bxHalfMrg),
				sf*(2*bxHalfMrg),
				sf*(2*bxHalfMrg),
			)
			c.SetColor(color.RGBA{32, 32, 32, 80})
			c.Fill()
		}

		if paling > 0.16 {
			c.DrawCircle(
				sf*x,
				sf*y,
				sf*rd,
			)
			c.SetColor(pale(col, paling))
			c.Fill()

			c.SetColor(pale(fontCol, paling))
			// c.SetRGBA(1, 1, 1, 1)
			c.DrawStringWrapped(
				sh,
				sf*x,
				sf*(y+bx/2),
				0.5, 0.99, // ax 0.5, ay 0.5,
				bx*0.95, // width
				1.3,
				gg.AlignCenter,
			)

		}

	}

	mainFrameCntr := -1 // number of images
	numSubFrames := 40.0
	for _, mainFrame := range mainFrames {
		mainFrameCntr++

		for subFrameCntr := 0.0; subFrameCntr < numSubFrames; subFrameCntr++ {
			share := subFrameCntr / numSubFrames

			// empty existing context from previous frame drawings
			c.SetRGB(0.001, 0.001, 0.001)
			c.Clear()

			for longName, itm := range mainFrame.Items {
				x := itm.X
				y := itm.Y
				rad := itm.Rad
				col := itm.Color
				isSubframe := subFrameCntr != 0 && mainFrameCntr != len(mainFrames)-1 // subframe and not last frame
				minRank := itm.Rank < 20 || itm.YearNext != nil && itm.YearNext.Rank < 20
				linearCombisPossible := itm.YearNext != nil // looking into future
				paling := 1.0
				if minRank && isSubframe {
					if linearCombisPossible {
						x = x + (itm.YearNext.X-x)*share
						y = y + (itm.YearNext.Y-y)*share
						rad = rad + (itm.YearNext.Rad-rad)*share
					} else if itm.YearNext == nil {
						// y = y + (110-y)*share
						y = y + share*(200-y) // fast slide
					} else if mainFrameCntr != 0 && itm.YearPrev == nil {
						y = y - share*(200-y) // fast rise
					}
				} else if isSubframe {
					paling = 0.7
					if share > 0.7 {
						paling = 1.0 - share
					}
				}

				drwC(
					x, y, rad,
					col,
					itm.Short,
					longName,
					itm.Box,
					paling,
				)
			}

			//
			// left top frame label
			yr := mainFrame.Year
			loadFont(c, 32)
			c.SetRGB(0.8, 0.8, 0.8)
			c.DrawString(fmt.Sprintf("Year %v  - %v-%v", yr, mainFrameCntr+1, subFrameCntr), 5, 5+c.FontHeight())
			loadFont(c, 12) // reset font

			//
			// save to PNG
			subset := mainFrameCntr == 0 || mainFrameCntr == len(mainFrames)-2 || mainFrameCntr == len(mainFrames)-3
			if subset || subFrameCntr == 0 {
				fn := fmt.Sprintf("./out/png/anim_trans_%02v_%02v.png", yr, subFrameCntr)
				c.SavePNG(fn)
			}

			// save to animated GIF structure
			images = append(images, renderIntoPalettedImage(c))
			elongation := 4
			if subFrameCntr == 0 {
				elongation = 550
				if mainFrameCntr == 0 {
					elongation = 1100
				}
				if mainFrameCntr == len(mainFrames)-1 {
					elongation = 22250
				}
			}
			delays = append(delays, elongation)

			log.Printf("mainfr,subfr %02v-%02v", mainFrameCntr, subFrameCntr)

		}

	}

	saveAnimGIF(images, delays, "anim_trans")

}

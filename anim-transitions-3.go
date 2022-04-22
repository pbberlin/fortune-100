package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

func animationsTransitionStage3() {

	mainFrames := Load("mainFrames2.json")

	//
	// output structures
	var images []*image.Paletted
	var delays []int

	// go graphics
	// 		we use a single context, cleaning it between frames
	c := gg.NewContext(int(w), int(h))
	loadFont(c, 12)

	sf := computeSF()

	drwC := func(
		x, y, rd float64,
		// companyRev float64,
		col color.Color,
		sh, longNammme string,
		bx float64,
	) {

		// drawing a pale box - to make the box packing easy to spot
		if true {
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

		c.DrawCircle(
			sf*x,
			sf*y,
			sf*rd,
		)
		c.SetColor(col)
		c.Fill()

		c.SetRGB(0.95, 0.95, 0.95)
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

	frameCntr := -1 // number of images
	for _, mainFrame := range mainFrames {
		frameCntr++

		yr := mainFrame.Year

		// empty existing context from previous frame drawings
		c.SetRGB(0.001, 0.001, 0.001)
		c.Clear()

		for longName, itm := range mainFrame.Items {

			// drawing a pale box - to make the box packing easy to spot
			// if true {
			// 	bxMrg := 0.1
			// 	c.DrawRectangle(sf*(x-bx/2+bxMrg), sf*(y+bxMrg), sf*(bx-2*bxMrg), sf*(bx-2*bxMrg))
			// 	c.SetColor(color.RGBA{32, 32, 32, 80})
			// 	c.Fill()
			// }

			drwC(
				itm.X, itm.Y, itm.Rad,
				// 2222,
				itm.Color,
				itm.Short,
				longName,
				itm.Box,
			)

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
			fn := fmt.Sprintf("./out/anim_trans_%02v.png", yr)
			c.SavePNG(fn)
		}

		// save to animated GIF structure
		images = append(images, renderIntoPalettedImage(c))
		elongation := 50
		delays = append(delays, elongation)

	}

	saveAnimGIF(images, delays, "anim_trans")

}

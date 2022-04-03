package main

import (
	"fmt"
	"image"
	"log"

	"github.com/fogleman/gg"
)

func stockMarket2(rankings []Ranking, companies []company, companiesByName map[string]company) {

	var images []*image.Paletted
	var delays []int

	var w, h float64
	w = 1024
	h = 768
	frames := 4

	c := gg.NewContext(int(w), int(h))
	if err := c.LoadFontFace("./out/arialbd.ttf", 96); err != nil {
		log.Fatalf("Cannot load font: %v", err)
	}
	//
	for frame := 0; frame < frames; frame++ {

		c.SetRGB(0, 0, 0)
		c.Clear()

		c.SetRGB(0.8, 0.8, 0.8)
		c.DrawString(fmt.Sprintf("Frame %v", frame), 5, 5+c.FontHeight())

		// save to PNG
		if true {
			fn := fmt.Sprintf("./out/sm2_%02v.png", frame)
			c.SavePNG(fn)
		}

		images = append(images, renderIntoPalettedImage(c))

		elongation := 10
		if frame < frames/2 {
			delays = append(delays, frame*elongation)
		} else {
			delays = append(delays, (frames-frame)*elongation)
		}

	}

	saveAnimGIF(images, delays, "stock-market-2")

}

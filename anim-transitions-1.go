package main

import (
	"encoding/json"
	"image/color"
	"log"
	"math"
	"os"
)

// structs for intediary stage

type Item struct {
	X, Y, Rad float64    // x,y coordinates 0...100 and circle radius
	Color     color.RGBA // company color
	Short     string     // name short
	Box       float64
}

type MainFrame struct {
	Year  int
	Items map[string]Item // company name as key
}

func animationsTransitionStage1() {

	// fetching data
	rksyrs := readRankingsYears()
	companiesByName := readCompaniesByYears()
	baseQuant := rksyrs.Qs[1] // quantile
	years := rksyrs.Years()

	//
	// intermediara structures
	mainFrames := []MainFrame{}

	addItem := func(
		fr *MainFrame,
		x, y, boxRad float64,
		companyRev float64,
		nameLong, nameShort string,
		col color.RGBA,
		box float64,
	) {
		zeroToOne := math.Sqrt(companyRev / baseQuant.Rev)
		cRad := zeroToOne * bxBaseRad // circle radius
		itm := Item{X: x, Y: (y + boxRad), Rad: cRad, Short: nameShort, Color: col}

		itm.X = math.Round(itm.X*1000) / 1000
		itm.Y = math.Round(itm.Y*1000) / 1000
		itm.Rad = math.Round(itm.Rad*1000) / 1000

		itm.Box = box

		fr.Items[nameLong] = itm

	}

	//
	//
	continuousRows := true // continuous rows - dont start new now on new quantile
	frameCntr := -1        // number of images
	for _, yr := range years {
		frameCntr++

		frame := MainFrame{Year: yr, Items: map[string]Item{}} // init
		mainFrames = append(mainFrames, frame)

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

			addItem(
				&frame,
				x, y, bx/2,
				rv,
				nm, sh,
				companiesByName[nm].Color,
				bx,
			)

		}

	}

	bts1, err := json.MarshalIndent(mainFrames, " ", "  ")
	if err != nil {
		log.Fatalf("cannot jsonify mainFrames: %v", err)
	}
	err = os.WriteFile("./out/mainFrames.json", bts1, 0777)
	if err != nil {
		log.Fatalf("cannot write file ./out/mainFrames.json: %v", err)
	}

}

package main

import (
	"encoding/json"
	"image/color"
	"log"
	"os"
	"path/filepath"
)

// structs for intediary stage

type ItemCore struct {
	X, Y, Rad   float64 // x,y coordinates 0...100 and circle radius
	Box         float64
	Long, Short string // name long, short; long is also the map key
	Rank        int
	Revenue     float64
	Color       color.RGBA // company color
}

type Item struct {
	ItemCore
	YearPrev *ItemCore // "linked list"
	YearNext *ItemCore
}

type MainFrame struct {
	Year  int
	Items map[string]Item // company name as key
}

type MainFrames []MainFrame

func Load(fn string) MainFrames {
	pth := filepath.Join(".", "out", fn)
	bts, err := os.ReadFile(pth)
	if err != nil {
		log.Fatalf("cannot read file %v: %v", fn, err)
	}

	mainFrames := MainFrames{}
	err = json.Unmarshal(bts, &mainFrames)
	if err != nil {
		log.Fatalf("cannot unmarshal %v: %v", fn, err)
	}
	return mainFrames
}

func (mfs MainFrames) Save(fn string) {
	pth := filepath.Join(".", "out", fn)
	bts1, err := json.MarshalIndent(mfs, " ", "  ")
	if err != nil {
		log.Fatalf("cannot jsonify mainFrames: %v", err)
	}
	err = os.WriteFile(pth, bts1, 0777)
	if err != nil {
		log.Fatalf("cannot write file %v: %v", pth, err)
	}

}

package main

import (
	"encoding/json"
	"image/color"
	"log"
	"math"
	"os"
	"sort"
	"strings"
)

// company is a helper struct
type company struct {
	Name  string
	Color color.RGBA
}

type Ranking struct {
	Year    int
	Rank    int
	Name    string
	Short   string // upper case
	Revenue float64
}

type RankingsYear struct {
	Year     int
	Min      float64 // Min revenue from all companies in this year
	Max      float64 // ...
	Rankings []Ranking
}

//
type RankingsYears struct {
	// redundant - same for each
	MinTtl float64 // across all years
	MaxTtl float64 // ...

	// various quantiles across all years
	//   i.e. Q[50] holds the median revenue - 50 percent
	// Qs map[int]float64
	Qs Quantils

	RkgsYear []RankingsYear
}

func (rksyrs RankingsYears) Years() []int {
	yrs := make([]int, 0, 10)
	last := -1
	for i := 0; i < len(rksyrs.RkgsYear); i++ {
		if rksyrs.RkgsYear[i].Year != last {
			last = rksyrs.RkgsYear[i].Year
			yrs = append(yrs, last)
		}
	}
	return yrs
}

func (rksyrs RankingsYears) Deflate(deflator float64) {
	yr0 := rksyrs.Years()[0]
	for i := 0; i < len(rksyrs.RkgsYear); i++ {
		// cumulative inflation
		cuml := math.Pow(deflator, float64(rksyrs.RkgsYear[i].Year-yr0))
		log.Printf("Deflator for yr %v is %4.2f  (%v**%2v)", rksyrs.RkgsYear[i].Year, cuml, deflator, rksyrs.RkgsYear[i].Year-yr0)
		for j := 0; j < len(rksyrs.RkgsYear[i].Rankings); j++ {
			rksyrs.RkgsYear[i].Rankings[j].Revenue /= cuml
		}
	}
}

func (rksyrs RankingsYears) SetMinMax() {
	for i := 0; i < len(rksyrs.RkgsYear); i++ {
		rksyrs.RkgsYear[i].Min = math.MaxFloat64
		rksyrs.RkgsYear[i].Max = -math.MaxFloat64
		for j := 0; j < len(rksyrs.RkgsYear[i].Rankings); j++ {
			if rksyrs.RkgsYear[i].Min > rksyrs.RkgsYear[i].Rankings[j].Revenue {
				rksyrs.RkgsYear[i].Min = rksyrs.RkgsYear[i].Rankings[j].Revenue
			}
			if rksyrs.RkgsYear[i].Max < rksyrs.RkgsYear[i].Rankings[j].Revenue {
				rksyrs.RkgsYear[i].Max = rksyrs.RkgsYear[i].Rankings[j].Revenue
			}
		}
		rksyrs.RkgsYear[i].Max = math.Ceil(rksyrs.RkgsYear[i].Max/1000) * 1000 // rounded by thousands
		rksyrs.RkgsYear[i].Min = math.Ceil(rksyrs.RkgsYear[i].Min/1000) * 1000
	}

}

func (rksyrs *RankingsYears) SetMinMaxTotal() {
	min := math.MaxFloat64
	max := -math.MaxFloat64
	for i := 0; i < len(rksyrs.RkgsYear); i++ {
		if min > rksyrs.RkgsYear[i].Min {
			min = rksyrs.RkgsYear[i].Min
		}
		if max < rksyrs.RkgsYear[i].Max {
			max = rksyrs.RkgsYear[i].Max
		}
	}
	rksyrs.MinTtl = math.Ceil(min/1000) * 1000 // rounded by thousands
	rksyrs.MaxTtl = math.Ceil(max/1000) * 1000
}

func (rksyrs RankingsYears) QuantilesTotal() {

	all := make([]Ranking, 0, len(rksyrs.RkgsYear)*103) // across all years
	for i := 0; i < len(rksyrs.RkgsYear); i++ {
		all = append(all, rksyrs.RkgsYear[i].Rankings...)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Revenue > all[j].Revenue {
			return false
		}
		return true
	})
	// dbg.Dump(all[0:5])
	// dbg.Dump(all[len(all)-5:])

	for qIdx := 0; qIdx < len(rksyrs.Qs); qIdx++ {
		if rksyrs.Qs[qIdx].Q == 0 || rksyrs.Qs[qIdx].Q == 100 {
			continue
		}
		idxQ := int(math.Floor(float64(len(all)) / 100 * float64(rksyrs.Qs[qIdx].Q)))
		rksyrs.Qs[qIdx].Rev = math.Floor(all[idxQ].Revenue/1000) * 1000
		rksyrs.Qs[qIdx].Idx = idxQ
		// log.Printf("quantile %3v has rank %3v of %v => %10.0f", rksyrs.Qs[qIdx].Q, rksyrs.Qs[qIdx].Idx, len(all), rksyrs.Qs[qIdx].Val)
	}

	rksyrs.Qs[000].Rev = rksyrs.MinTtl
	rksyrs.Qs[len(rksyrs.Qs)-1].Rev = rksyrs.MaxTtl
	rksyrs.Qs[len(rksyrs.Qs)-1].Idx = len(all) - 1

	for _, quant := range rksyrs.Qs {
		log.Printf("quantile %3v has rank %3v of %v => %10.0f", quant.Q, quant.Idx+1, len(all), quant.Rev)
	}

}

func (rksyrs RankingsYears) NamesShort() {
	for i := 0; i < len(rksyrs.RkgsYear); i++ {
		for j := 0; j < len(rksyrs.RkgsYear[i].Rankings); j++ {
			sh := ""
			for _, rn := range rksyrs.RkgsYear[i].Rankings[j].Name {
				s := string(rn)
				if s == " " {
					continue
				}
				// extract upper case letters
				if s == strings.ToUpper(s) {
					sh += s
				}
			}
			if len(sh) < 2 {
				sh = rksyrs.RkgsYear[i].Rankings[j].Name[:3]
			}
			if len(sh) > 4 {
				sh = rksyrs.RkgsYear[i].Rankings[j].Name[:5]
			}
			rksyrs.RkgsYear[i].Rankings[j].Short = sh

			if shExp, ok := explicitShortNames[rksyrs.RkgsYear[i].Rankings[j].Name]; ok {
				rksyrs.RkgsYear[i].Rankings[j].Short = shExp
			}

		}
	}
}

// organize hierarchifies flat rankings by year
func organize() {

	bts1, err := os.ReadFile("./out/rankings.json")
	if err != nil {
		log.Fatalf("cannot read file ./out/rankings.json: %v", err)
	}

	rankings := []Ranking{}
	err = json.Unmarshal(bts1, &rankings)
	if err != nil {
		log.Fatalf("cannot unmarshal ./out/rankings.json: %v", err)
	}

	sort.Slice(rankings, func(i, j int) bool {
		if rankings[i].Year > rankings[j].Year {
			return false
		}
		if rankings[i].Year < rankings[j].Year {
			return true
		}
		// year equality
		if rankings[i].Rank < rankings[j].Rank {
			return false
		}
		return true
	})

	// dbg.Dump(rankings[:4])
	// dbg.Dump(rankings[98:102])
	// dbg.Dump(rankings[len(rankings)-4:])

	// new structure - rankings by years
	last := -1
	rksyrs := RankingsYears{}
	// init map for quantiles
	rksyrs.Qs = Quantils{
		{Q: 00},
		{Q: 50},
		{Q: 75},
		{Q: 90},
		{Q: 95},
		// {Q: 97},
		{Q: 98},
		{Q: 100},
	}

	rksyrs.Qs = Quantils{
		{Q: 00},
		{Q: 20},
		{Q: 40},
		{Q: 60},
		{Q: 80},
		{Q: 90},
		{Q: 93},
		{Q: 95},
		{Q: 97},
		{Q: 98},
		{Q: 99},
		{Q: 100},
	}

	var rksYear RankingsYear
	for i := 0; i < len(rankings); i++ {
		if rankings[i].Year != last {
			if last > -1 {
				rksyrs.RkgsYear = append(rksyrs.RkgsYear, rksYear)
			}
			last = rankings[i].Year
			// init new
			rksYear = RankingsYear{}
			rksYear.Year = rankings[i].Year
			rksYear.Rankings = []Ranking{}
		}
		rksYear.Rankings = append(rksYear.Rankings, rankings[i])
	}
	rksyrs.RkgsYear = append(rksyrs.RkgsYear, rksYear)

	rksyrs.NamesShort()
	rksyrs.Deflate(1.025) // cautious
	rksyrs.SetMinMax()
	rksyrs.SetMinMaxTotal()
	rksyrs.QuantilesTotal()

	bts2, err := json.MarshalIndent(rksyrs, " ", "  ")
	if err != nil {
		log.Fatalf("cannot jsonify rksYears: %v", err)
	}
	err = os.WriteFile("./out/rksYears.json", bts2, 0777)
	if err != nil {
		log.Fatalf("cannot write file ./out/rksYears.json: %v", err)
	}

	//
	//
	// distinct companies
	companies := make([]company, 0, 100)
	companiesByName := map[string]company{}
	for i := 0; i < len(rankings); i++ {
		if _, ok := companiesByName[rankings[i].Name]; !ok {
			companies = append(companies, company{Name: rankings[i].Name})
		}
		companiesByName[rankings[i].Name] = company{Name: rankings[i].Name}
	}

	// allocate colors
	for i := 0; i < len(companies); i++ {
		companies[i].Color = circleCols[(i % len(circleCols))]
		companiesByName[companies[i].Name] = companies[i]
	}

	log.Printf("Found %v distinct companies", len(companies))

	bts3, err := json.MarshalIndent(companiesByName, " ", "  ")
	if err != nil {
		log.Fatalf("cannot jsonify companiesByName: %v", err)
	}
	err = os.WriteFile("./out/companiesByName.json", bts3, 0777)
	if err != nil {
		log.Fatalf("cannot write file ./out/companiesByName.json: %v", err)
	}

	// dbg.Dump(companies[:4])
	// dbg.Dump(companies[len(companies)-4:])

}

func readRankingsYears() RankingsYears {

	bts, err := os.ReadFile("./out/rksYears.json")
	if err != nil {
		log.Fatalf("cannot read file ./out/rksYears.json: %v", err)
	}

	rankingsYears := RankingsYears{}
	err = json.Unmarshal(bts, &rankingsYears)
	if err != nil {
		log.Fatalf("cannot unmarshal ./out/rksYears.json: %v", err)
	}

	return rankingsYears
}

func readCompaniesByYears() map[string]company {

	bts, err := os.ReadFile("./out/companiesByName.json")
	if err != nil {
		log.Fatalf("cannot read file ./out/companiesByName.json: %v", err)
	}

	companiesByName := map[string]company{}
	err = json.Unmarshal(bts, &companiesByName)
	if err != nil {
		log.Fatalf("cannot unmarshal ./out/companiesByName.json: %v", err)
	}

	return companiesByName
}

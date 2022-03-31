package main

import (
	"encoding/json"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pbberlin/dbg"
)

type Ranking struct {
	Year    int
	Rank    int
	Name    string
	Revenue int
}

func prepareData() {

	rankings := []Ranking{}

	fs, err := os.ReadDir("./raw/")
	if err != nil {
		log.Fatalf("cannot read files raw data: %v", err)
	}
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		log.Printf("processing %v", f.Name())
		bts, err := os.ReadFile(filepath.Join("./raw", f.Name()))
		if err != nil {
			log.Fatalf("cannot read content of %v: %v", f.Name(), err)
		}
		cnt := string(bts)
		lines := strings.Split(cnt, "\r\n")

		if len(lines) < 10 {
			log.Fatalf("must be something wrong with fileconten - splitting by \\r\\n yielded no lines %v", f.Name())
		}

		startStop := []int{}
		for i := 0; i < len(lines); i++ {
			if lines[i] == "1" {
				log.Printf("   %v: start/stop at %3v - %q", f.Name(), i, lines[i])
				startStop = append(startStop, i)
			}
		}

		if len(startStop) != 2 {
			log.Printf("no start-stop positions found: %v", f.Name())
			continue
		}
		pos2 := startStop[1]
		if lines[pos2-1] != "Page " {
			log.Fatalf("second '1' must be preceded by 'Page '; is %q", lines[pos2-1])
		}
		startStop[1]--

		for i := startStop[0]; i < startStop[1]; i += 3 {
			rnk := Ranking{}
			yr := filepath.Base(f.Name())
			yr = strings.TrimSuffix(yr, filepath.Ext(yr))
			rnk.Year, err = strconv.Atoi(yr)
			if err != nil {
				log.Printf("cannot get the int from %v: %v", yr, err)
			}
			rnk.Rank, err = strconv.Atoi(lines[i])
			if err != nil {
				log.Printf("cannot get the int from %v: %v", lines[i], err)
			}
			rnk.Name = lines[i+1]

			rev := lines[i+2]
			if strings.HasPrefix(rev, "$") {
				rev = rev[1:]
			}
			rev = strings.ReplaceAll(rev, ",", "")
			if pos := strings.Index(rev, "."); pos > -1 {
				rev = rev[:pos]
			}
			rnk.Revenue, err = strconv.Atoi(rev)
			if err != nil {
				log.Printf("cannot get the int from %v: %v", rev, err)
			}
			rankings = append(rankings, rnk)
		}

	}

	bts, err := json.MarshalIndent(rankings, " ", "  ")
	if err != nil {
		log.Fatalf("cannot jsonify rankings: %v", err)
	}
	err = os.WriteFile("./out/rankings.json", bts, 0777)
	if err != nil {
		log.Fatalf("cannot write file ./out/rankings.json: %v", err)
	}

}

type company struct {
	Name  string
	Color color.RGBA
}

func restruct() []Ranking {

	bts, err := os.ReadFile("./out/rankings.json")
	if err != nil {
		log.Fatalf("cannot read file ./out/rankings.json: %v", err)
	}

	rankings := []Ranking{}
	err = json.Unmarshal(bts, &rankings)
	if err != nil {
		log.Fatalf("cannot unmarshal ./out/rankings.json: %v", err)
	}

	dbg.Dump(rankings[:4])

	// distinct companies
	companies := make([]company, 0, 100)
	distinct := map[string]interface{}{}
	for i := 0; i < len(rankings); i++ {
		if _, ok := distinct[rankings[i].Name]; ok {
			companies = append(companies, company{Name: rankings[i].Name})
		}
		distinct[rankings[i].Name] = nil
	}
	log.Printf("Found %v distinct companies", len(companies))

	// allocate colors
	var cols = []color.RGBA{
		// {0x00, 0x00, 0x00, 0xff}, // not black
		{0x00, 0x00, 0xff, 0xff},
		{0x00, 0xff, 0x00, 0xff},
		{0x00, 0xff, 0xff, 0xff},
		{0xff, 0x00, 0x00, 0xff},
		{0xff, 0x00, 0xff, 0xff},
		{0xff, 0xff, 0x00, 0xff},
		{0xff, 0xff, 0xff, 0xff},
	}
	for i := 0; i < len(companies); i++ {
		companies[i].Color = cols[(i % len(cols))]
	}

	dbg.Dump(companies[:4])
	dbg.Dump(companies[len(companies)-4:])

	return rankings

}

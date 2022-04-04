package main

import (
	"encoding/json"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Ranking struct {
	Year    int
	Rank    int
	Name    string
	Revenue int
}

type RankingsYear struct {
	Year     int
	Rankings []Ranking
}

type RankingsYears []RankingsYear

func (rksyrs RankingsYears) Years() []int {
	yrs := make([]int, 0, 10)
	last := -1
	for i := 0; i < len(rksyrs); i++ {
		if rksyrs[i].Year != last {
			last = rksyrs[i].Year
			yrs = append(yrs, last)
		}
	}
	return yrs
}

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

// company is a helper struct
type company struct {
	Name  string
	Color color.RGBA
}

var suffixes1 = []string{
	"& company",
	"company",
	"companies",
	"comp",
	"corporation",
	"corp",
	"corp.",
	"& inc.",
	"inc.",
	"and co.",
	"& co.",
	"co.",
	"group",
	"insurance",
	"holdings",
}

var suffixes2 = []string{
	", ",
	",",
	" ",
}

func cleanseName(s string) string {
	sl := strings.ToLower(s)
	// log.Printf("Name cleanse %v", sl)
	if strings.HasPrefix(sl, "the ") {
		log.Printf("  cleanse %q from %q", "the ", s)
		s = s[4:]
	}
	for _, sfx1 := range suffixes1 {
		for _, sfx2 := range suffixes2 {
			sfx := sfx2 + sfx1
			sl := strings.ToLower(s)
			if strings.HasSuffix(sl, sfx) {
				log.Printf("  cleanse %q from %q", sfx, s)
				s = s[:len(s)-len(sfx)]
			}
		}

	}

	return s
}

var reps = map[string]string{
	"Morgan Stanley Dean Witter":       "Morgan Stanley",
	"MCI WorldCom":                     "MCI",
	"MCI Communications":               "MCI",
	"Express Scripts Holding":          "Express Scripts",
	"Aetna Life & Casualty":            "Aetna",
	"International Business Machines":  "IBM",
	"Hewlett-Packard":                  "HP",
	"Walgreens Boots Alliance":         "Walgreen",
	"Walgreens":                        "Walgreen",
	"ConocoPhillips":                   "Conoco",
	"HCA Inc":                          "HCA",
	"HCA Healthcare":                   "HCA",
	"Columbia/HCA Healthcare":          "HCA",
	"CVS Health":                       "CVS",
	"CVS Caremark":                     "CVS",
	"Travelers Cos.":                   "Travelers",
	"Costco Wholesale":                 "Costco",
	"Sprint Nextel":                    "Sprint",
	"Price/Costco":                     "Costco",
	"State Farm Insurance Cos.":        "State Farm",
	"Amazon.com":                       "Amazon",
	"Dell Technologies":                "Dell",
	"Dell Computer":                    "Dell",
	"United Parcel Service of America": "United Parcel Service",
	"McKesson HBOC":                    "McKesson",
	"Prudential Financial (U.S.)":      "Prudential Financial",
	"Verizon Communications":           "Verizon",
	"Philip Morris International":      "Philip Morris",
	"Raytheon Technologies":            "Raytheon",
	"Dow Chemical":                     "Dow",
	"UnitedHealth Group Incorporated":  "UnitedHealth",
	"UnitedHealth Group, Incorporated": "UnitedHealth",
	"St. Paul Travelers":               "Travelers",
	"Nationwide Mutual":                "Nationwide",
	"Liberty Mutual Holding":           "Liberty Mutual",
	"Sears, Roebuck and":               "Sears",
	"Sears, Roebuck":                   "Sears",
	"J.P. Morgan Chase":                "J.P. Morgan",
	"J.P. Morgan & Co. Incorporated":   "J.P. Morgan",
	"E.I. du Pont de Nemours and":      "E.I. du Pont de Nemours",
	"TIAA-CREF":                        "TIAA",
	"Energy Transfer Equity":           "Energy Transfer",
}

func replaceName(s string) string {
	s = strings.TrimSpace(s)
	if s1, ok := reps[s]; ok {
		s = s1
	}
	return s
}

func rawList2JSON() {

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
			rnk.Name = cleanseName(rnk.Name)
			rnk.Name = cleanseName(rnk.Name)
			rnk.Name = replaceName(rnk.Name)

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

	for i := 0; i < len(rankings); i++ {
		for j := 0; j < len(rankings); j++ {
			if rankings[i].Name == rankings[j].Name {
				continue
			}
			if strings.Contains(rankings[i].Name, rankings[j].Name) {
				log.Printf("%q part of %q", rankings[i].Name, rankings[j].Name)
			}
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

func organize() ([]Ranking, RankingsYears, []company, map[string]company) {

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

	last := -1
	rksYears := RankingsYears{}
	var rksYear RankingsYear
	for i := 0; i < len(rankings); i++ {
		if rankings[i].Year != last {
			if last > -1 {
				rksYears = append(rksYears, rksYear)
			}
			last = rankings[i].Year
			// init new
			rksYear = RankingsYear{}
			rksYear.Year = rankings[i].Year
			rksYear.Rankings = []Ranking{}
		}
		rksYear.Rankings = append(rksYear.Rankings, rankings[i])
	}
	rksYears = append(rksYears, rksYear)

	bts2, err := json.MarshalIndent(rksYears, " ", "  ")
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

	for i := 0; i < len(companies); i++ {
		companies[i].Color = cols[(i % len(cols))]
		companiesByName[companies[i].Name] = companies[i]
		// log.Printf("%4v %-22v - %v  - col %v", i, companies[i].Name, (i % len(cols)), companies[i].Color)
	}

	log.Printf("Found %v distinct companies", len(companies))
	log.Printf("Found %v distinct companies", len(companiesByName))

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

	return rankings, rksYears, companies, companiesByName

}

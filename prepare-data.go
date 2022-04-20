package main

import (
	"encoding/json"
	"image/color"
	"log"
	"math"
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
	Short   string // upper case
	Revenue float64
}

type RankingsYear struct {
	Year     int
	Min      float64 // Min revenue from all companies in this year
	Max      float64 // ...
	Rankings []Ranking
}

type Quantil struct {
	Q   int // 50 percent, or 90 percent...
	Rev float64

	Idx int // index in the base set, i.e. Q90 is 540 out of 600
}

type Quantils []Quantil // we want them ordered, therefore not as map

func (qs Quantils) At(q int) Quantil {
	for i := 0; i < len(qs); i++ {
		if q == qs[i].Q {
			return qs[i]
		}
	}
	log.Panicf("quantile %v not set", q)
	return Quantil{}
}

func (qs Quantils) Next(q int) Quantil {
	for i := 0; i < len(qs); i++ {
		if q == qs[i].Q {
			if i != len(qs)-1 {
				return qs[i+1]
			}
			log.Printf("quantile %v is last; returning last.", i)
			return qs[len(qs)-1]
		}
	}
	log.Panicf("quantile next of %v does not exit", q)
	return Quantil{}
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

var circleCols = []color.RGBA{
	// {0x00, 0x00, 0x00, 0xff}, // not black

	{0, 0, 192, 255},
	{0, 0, 128, 255},

	{0, 192, 0, 255},
	{0, 128, 0, 255},

	{0, 192, 192, 255},
	{0, 192, 128, 255},
	{0, 128, 192, 255},
	{0, 128, 128, 255},

	{192, 0, 0, 255},
	{128, 0, 0, 255},

	{192, 0, 192, 255},
	{192, 0, 128, 255},
	{128, 0, 192, 255},
	{128, 0, 128, 255},

	{192, 192, 0, 255},
	{192, 128, 0, 255},
	{128, 192, 0, 255},
	{128, 128, 0, 255},

	{128, 128, 128, 255}, // for get Walmart != Amazon

	// {0xff, 0xff, 0xff, 255}, // not white
}

// company is a helper struct
type company struct {
	Name  string
	Color color.RGBA
}

var suffixes1 = []string{
	", ",
	",",
	" ",
}

var suffixes2 = []string{
	"& company",
	"company",
	"companies",
	"comp",
	"corporation",
	"corp",
	"corp.",
	"incorporated",
	"& inc.",
	"inc.",
	"and co.",
	"& co.",
	"co.",
	"international",
	"group",
	"insurance",
	"holding",
	"holdings",
}

func cleanseName(s string) string {

	sl := strings.ToLower(s)
	// log.Printf("Name cleanse %v", sl)

	if strings.HasPrefix(sl, "the ") {
		s = s[4:]
	}
	if strings.HasPrefix(sl, "minnesota mining") {
		s = "Minnesota Mining"
	}
	for _, sfx1 := range suffixes2 {
		for _, sfx2 := range suffixes1 {
			sfx := sfx2 + sfx1
			sl := strings.ToLower(s)
			if strings.HasSuffix(sl, sfx) {
				// log.Printf("  cleanse %q from %q", sfx, s)
				s = s[:len(s)-len(sfx)]
			}
		}
	}

	return s
}

var replacements = map[string]string{
	"Wal-Mart Stores": "Walmart",

	"American International Group,Inc.": "AIG",
	"American International Group":      "AIG",
	"Enterprise GP Holdings, L.P.":      "Enterpr GP Hd",
	"International Assets":              "Intl Assets",
	"International Assets Holding":      "Intl Assets",

	"Marathon Petroleum":                      "Marathon Petrol.",
	"Phillips Petroleum":                      "Phillips Petrol.",
	"Metropolitan Life":                       "Metropol. Life",
	"Morgan Stanley Dean Witter":              "Morgan Stanley",
	"MCI WorldCom":                            "MCI",
	"MCI Communications":                      "MCI",
	"Express Scripts Holding":                 "Express Scripts",
	"Aetna Life & Casualty":                   "Aetna",
	"International Business Machines":         "IBM",
	"Hewlett-Packard":                         "HP",
	"Walgreens Boots Alliance":                "Walgreen",
	"Walgreens":                               "Walgreen",
	"ConocoPhillips":                          "Conoco",
	"HCA Inc":                                 "HCA",
	"HCA Healthcare":                          "HCA",
	"Columbia/HCA Healthcare":                 "HCA",
	"CVS Health":                              "CVS",
	"CVS Caremark":                            "CVS",
	"Travelers Cos.":                          "Travelers",
	"Costco Wholesale":                        "Costco",
	"Sprint Nextel":                           "Sprint",
	"Price/Costco":                            "Costco",
	"State Farm Insurance Cos.":               "State Farm",
	"Amazon.com":                              "Amazon",
	"Dell Technologies":                       "Dell",
	"Dell Computer":                           "Dell",
	"McKesson HBOC":                           "McKesson",
	"Verizon Communications":                  "Verizon",
	"Philip Morris International":             "Philip Morris",
	"Raytheon Technologies":                   "Raytheon",
	"Dow Chemical":                            "Dow",
	"UnitedHealth Group Incorporated":         "UnitedHealth",
	"UnitedHealth Group, Incorporated":        "UnitedHealth",
	"St. Paul Travelers":                      "Travelers",
	"Nationwide Mutual":                       "Nationwide",
	"Liberty Mutual Holding":                  "Liberty Mutual",
	"Sears, Roebuck and":                      "Sears",
	"Sears, Roebuck":                          "Sears",
	"J.P. Morgan Chase":                       "J.P. Morgan",
	"J.P. Morgan & Co. Incorporated":          "J.P. Morgan",
	"E.I. du Pont de Nemours and":             "DuPont",
	"E.I. Du Pont de Nemours and":             "DuPont",
	"E.I. du Pont de Nemours":                 "DuPont",
	"TIAA-CREF":                               "TIAA",
	"Energy Transfer Equity":                  "Energy Transfer",
	"International Paper":                     "Intl Paper",
	"Honeywell International":                 "Honeywell",
	"American Express":                        "AmEx",
	"FleetBoston Financial":                   "FleetBoston",
	"UtiliCorp United":                        "UtiliCorp",
	"Electronic Data Systems":                 "EDS",
	"Federated Depart\u0004ment Stores":       "Fed Dpt Stores",
	"American International":                  "American Int",
	"United Technologies":                     "United Tech",
	"United Parcel Service of America":        "UPS",
	"United Parcel Service":                   "UPS",
	"Medco Health Solutions":                  "Medco",
	"Prudential Insurance Company of America": "Prudential",
	"Prudential Financial (U.S.)":             "Prudential",
	"Prudential Financial":                    "Prudential",
	"Rockwell International":                  "Rockwell",
	"American Home Products":                  "American Home",
	"Goodyear Tire \u0026 Rubber":             "Goodyear",
	"Federal National Mortgage Association":   "Fed Mortgage",
	"Teachers Insurance and Annuity Association College Retiremen": "Teachers Insurance",
	"Mondelez International":             "Mondelez",
	"Enterprise Products Partners":       "EP Partners",
	"Massachusetts Mutual Life":          "Mass Mutual",
	"Massachusetts Financial":            "Mass Financial",
	"Plains All American Pipeline, L.P.": "Plains Pipeline",
	"Twenty-First Century Fox":           "21 Century Fox",
	"Archer-Daniels-Midland":             "ADL",
	"Archer Daniels Midland":             "ADL",
	"Thermo Fisher Scientific":           "Thermo Fisher",
	"Abbott Laboratories":                "Abbott",
	"Capital One Financial":              "Capital One",
	"Hartford Financial Services":        "Hartford Fin",
	"Charter Communications":             "Charter Commu",
	"Publix Super Markets":               "Publix Markets",
	"World Fuel Services":                "World Fuel",
	"Lucent Technologies":                "Lucent",

	"Minnesota Mining & Manufacturing ": "Minnesota Mining",
	"Minnesota Mining & Manufacturing":  "Minnesota Mining",
	"May Department Stores":             "May Dept Stores",
	"SBC Communications":                "SBC Comm",
	"AmerisourceBergen":                 "Amerisource Bergen",
}

var explicitShortNames = map[string]string{
	"Ford Motor":         "Ford",
	"Amazon":             "Amazn",
	"Walmart":            "Walm",
	"Exxon Mobil":        "Exxon",
	"Chevron":            "Chevrn",
	"Citicorp":           "Citi",
	"PepsiCo":            "Pepsi",
	"Bank of America":    "BkOfA",
	"Conoco":             "Conoco",
	"Apple":              "Apple",
	"Berkshire Hathaway": "Bksh.H",
}

func replaceName(s string) string {
	s = strings.TrimSpace(s)
	if s1, ok := replacements[s]; ok {
		s = s1
	}
	return s
}

// rawList2JSON parses string to numbers,
// removes 'comp.', ', inc.' in dozens of variations - a kind of 'stemming'
// normalizes company names dozens of namings
// saves into a flat list of rankings at ./out/rankings.json
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
			rnk.Name = replaceName(rnk.Name)
			rnk.Name = cleanseName(rnk.Name)
			rnk.Name = cleanseName(rnk.Name)
			rnk.Name = replaceName(rnk.Name)

			rnk.Name = strings.ReplaceAll(rnk.Name, "-", " ") // for line break

			if len(rnk.Name) > 16 {
				log.Printf("too long %v", rnk.Name)
			}

			rev := lines[i+2]
			if strings.HasPrefix(rev, "$") {
				rev = rev[1:]
			}
			rev = strings.ReplaceAll(rev, ",", "")
			if pos := strings.Index(rev, "."); pos > -1 {
				rev = rev[:pos]
			}
			revInt, err := strconv.Atoi(rev)
			if err != nil {
				log.Printf("cannot get the int from %v: %v", revInt, err)
			}
			rnk.Revenue = float64(revInt)
			rankings = append(rankings, rnk)
		}

	}

	for i := 0; i < len(rankings); i++ {
		for j := 0; j < len(rankings); j++ {
			if rankings[i].Name == rankings[j].Name {
				continue
			}
			if strings.Contains(rankings[i].Name, rankings[j].Name) {
				log.Printf("%16q part of %24q", rankings[j].Name, rankings[i].Name)
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
		{Q: 95},
		{Q: 97},
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

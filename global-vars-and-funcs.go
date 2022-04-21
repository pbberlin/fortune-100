package main

import (
	"encoding/json"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
)

var w float64 = 1024
var h float64 = 768

// initial box size,
var bxBase = 5.54 // 133 / 7.8   => roughly 17
// var bxBase = 7.35 // 18
// var bxBase = 6.63 // 20
// var bxBase = 6.04 // 22

var bxBaseRad = bxBase / 2

// all rendering arguments are standardized to
//   100 units of canvas height;
//   thus, 133.3 is the according max width
// wOverH := 1024 * 100.0 / 768 // width over height
var wOverH = w * 100.0 / h // width over height

// The context should be populated in
// coordinates from 0...100 and 0...133 - depending on the ratio could be 0...125 or other.
// If w and or h are changed, then the program should not be affected.
// Now a scale factor (sf) assuming min...max is 0...100;
// 		which is applied ot all painting operations immedatialy before painting
func computeSF() float64 {
	sf := float64(w) / float64(100)
	if h < w {
		sf = float64(h) / float64(100) // shorter side dominates
	}
	return sf
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
	"Exxon":              "Exxon",
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

func loadFont(c *gg.Context, fontSize float64) {
	// fontSize := 96.0
	// fontSize = 12.0
	if err := c.LoadFontFace("./out/arial.ttf", fontSize); err != nil {
		log.Fatalf("Cannot load font: %v", err)
	}

}

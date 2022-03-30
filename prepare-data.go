package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type comp struct {
	Year    int
	Rank    int
	Name    string
	Revenue int
}

func prepareData() {

	comps := []comp{}

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
			comp := comp{}
			yr := filepath.Base(f.Name())
			yr = strings.TrimSuffix(yr, filepath.Ext(yr))
			comp.Year, err = strconv.Atoi(yr)
			if err != nil {
				log.Printf("cannot get the int from %v: %v", yr, err)
			}
			comp.Rank, err = strconv.Atoi(lines[i])
			if err != nil {
				log.Printf("cannot get the int from %v: %v", lines[i], err)
			}
			comp.Name = lines[i+1]

			rev := lines[i+2]
			if strings.HasPrefix(rev, "$") {
				rev = rev[1:]
			}
			rev = strings.ReplaceAll(rev, ",", "")
			if pos := strings.Index(rev, "."); pos > -1 {
				rev = rev[:pos]
			}
			comp.Revenue, err = strconv.Atoi(rev)
			if err != nil {
				log.Printf("cannot get the int from %v: %v", rev, err)
			}
			comps = append(comps, comp)
		}

	}

	bts, err := json.MarshalIndent(comps, " ", "  ")
	if err != nil {
		log.Fatalf("cannot jsonify comps: %v", err)
	}
	err = os.WriteFile("./out/comps.json", bts, 0777)
	if err != nil {
		log.Fatalf("cannot write file ./out/comps.json: %v", err)
	}

}

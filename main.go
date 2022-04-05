package main

import (
	"log"
)

func main() {

	log.SetFlags(log.Lshortfile)

	if false {
		bezier()
		animGIF1()
		animGIF2()
	}

	// rawList2JSON()

	rankings, rksYears, companies, compantiesByName := organize()

	_, _ = rankings, companies
	stockMarket2(rksYears, compantiesByName)

}

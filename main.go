package main

import (
	"log"
)

func main() {

	log.SetFlags(log.Lshortfile)

	prepareData()

	bezier()
	animGIF1()
	animGIF2()
	stockMarket2()

}

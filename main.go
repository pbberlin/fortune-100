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

	// organize()

	// animationsNoTransition()

	animationsTransitionStage1()
	computeTransitions()
	animationsTransitionStage3()

}

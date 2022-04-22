package main

import (
	"log"
)

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

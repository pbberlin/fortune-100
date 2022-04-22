package main

func computeTransitions() {

	mainFrames := Load("mainFrames1.json")

	//
	for i1 := 0; i1 < len(mainFrames); i1++ {
		// yr := mainFrames[i1].Year
		// _ = yr
		items := mainFrames[i1].Items

		// next
		if i1 == len(mainFrames)-1 {
			continue // last year has no next
		}
		mfNextItems := mainFrames[i1+1].Items
		for lg, _ := range items {
			successor, ok := mfNextItems[lg]
			if ok {
				itm := items[lg]
				itm.YearNext = &successor.ItemCore
				mainFrames[i1].Items[lg] = itm
			}
		}

		// previous
		if i1 == 0 {
			continue // first year has no prev
		}
		mfPrevItems := mainFrames[i1-1].Items
		for lg, _ := range items {
			successor, ok := mfPrevItems[lg]
			if ok {
				itm := items[lg]
				itm.YearPrev = &successor.ItemCore
				mainFrames[i1].Items[lg] = itm
			}
		}

	}

	//
	//

	mainFrames.Save("mainFrames2.json")

}

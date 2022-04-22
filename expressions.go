package main

// ExpressionExecutor Математика пользовательских выражений
func ExpressionExecutor(price *PriceTerminal) {
	App.LockerPrices.Lock()
	App.LockerMath.Lock()

	for _, exp := range App.Math {
		var A *PriceTerminal
		var B *PriceTerminal
		var O *PriceTerminal


		if _, ok := App.Prices[price.Provider]; !ok {
			continue
		}

		A, okA := App.Prices[price.Provider][exp.ASeccode]

		if !okA {
			continue
		}

		B, okB := App.Prices[price.Provider][exp.BSeccode]

		if !okB {
			continue
		}

		O = App.Prices[price.Provider][exp.Out]

		// Тут меняем если есть
		if O != nil {
			O.Low = A.Low * B.Low       //priceLinked.Low   * 5469
			O.Close = A.Close * B.Close //priceLinked.Close * 5469
			O.High = A.High * B.High    //priceLinked.High  * 5469
			O.Open = A.High * B.Open    //priceLinked.High  * 5469
			O.DateTime = price.DateTime //priceLinked.High  * 5469
		} else {
			O = &PriceTerminal{
				SecCode:   exp.Out,
				DateTime:  price.DateTime,
				Close:     A.Close * B.Close,
				Percent:   price.Percent,
				Go:        price.Go,
				Open:      A.Open * B.Open,
				High:      A.High * B.High,
				Low:       A.Low * B.Low,
				Vol:       price.Vol,
				H5:        price.H5,
				L5:        price.L5,
				PriceStep: price.PriceStep,
				StepPrice: price.StepPrice,
				Provider:  price.Provider,
			}

			App.Prices[price.Provider][O.SecCode] = O
		}
	}


	App.LockerPrices.Unlock()
	App.LockerMath.Unlock()
}

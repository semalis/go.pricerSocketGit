package main

import (
	"JH-socket/Telega"
	"JHK-socket/Configs"
	"LK-socket/TCP"
	"log"
	"runtime/debug"
	"strconv"
)

func onMessage(client *TCP.Client, msg []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered:", r)
			debug.PrintStack()
		}
	}()


	//timeStart := time.Now()

	//log.Printf("Income Message: %s", msg)

	data := string(msg)

	price := PriceTerminal{} // то есть вот это делать ) если уже есть то проще в существующем поменять

	var prev int

	for z := 0; z <= len(data); z++ {
		if z == len(data) || data[z] == ':' {
			substr := data[prev:z]

			for e := 0; e < len(substr); e++ {
				if substr[e] == '=' {
					aS := substr[:e]
					bS := substr[e+1:]

					switch aS {

					// Запоминаем время теминала
					case "terminalName":
						App.SetTerminalAlive(bS)
						return

					case "seccode":
						price.SecCode = bS

					case "datetime":
						v, err := strconv.ParseInt(bS, 10, 64)

						if err != nil {
							log.Println("Error: Parse datetime:", err)
							continue
						}

						price.DateTime = v

					case "close":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse close:", err)
							continue
						}


						price.Close = v

					case "percent":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse percent:", err)
							continue
						}

						price.Percent = v

					case "go":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse go:", err)
							continue
						}

						price.Go = v


					case "open":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse open:", err)
							continue
						}

						price.Open = v


					case "high":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse high:", err)
							continue
						}

						price.High = v

					case "low":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse low:", err)
							continue
						}

						price.Low = v

					case "vol":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse vol:", err)
							continue
						}

						price.Vol = v

					case "h5":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse h5:", err)
							continue
						}

						price.H5 = v

					case "l5":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse l5:", err)
							continue
						}

						price.L5 = v

					case "price_step":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse price_step:", err)
							continue
						}

						price.PriceStep = v

					case "step_price":
						v, err := strconv.ParseFloat(bS, 64)

						if err != nil {
							log.Println("Error: Parse step_price:", err)
							continue
						}

						price.StepPrice = v

					case "provider":
						price.Provider = bS
					}
				}
			}

			z++
			prev = z
		}
	}

	/*

		for _, val := range strings.Split(data, ":") {
			row := strings.Split(val, "=")

			if len(row) != 2 {
				continue
			}

			switch row[0] {

			// Запоминаем время теминала
			case "terminalName":
				App.SetTerminalAlive(row[1])
				return

			case "seccode":
				price.SecCode = row[1]

			case "datetime":
				v, err := strconv.ParseInt(row[1], 10, 64)

				if err != nil {
					log.Println("Error: Parse datetime:", err)
					continue
				}

				price.DateTime = v

			case "close":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse close:", err)
					continue
				}

				price.Close = v

			case "percent":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse percent:", err)
					continue
				}

				price.Percent = v

			case "go":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse go:", err)
					continue
				}

				price.Go = v

			case "open":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse open:", err)
					continue
				}

				price.Open = v

			case "high":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse high:", err)
					continue
				}

				price.High = v

			case "low":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse low:", err)
					continue
				}

				price.Low = v

			case "vol":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse vol:", err)
					continue
				}

				price.Vol = v

			case "h5":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse h5:", err)
					continue
				}

				price.H5 = v

			case "l5":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse l5:", err)
					continue
				}

				price.L5 = v

			case "price_step":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse price_step:", err)
					continue
				}

				price.PriceStep = v

			case "step_price":
				v, err := strconv.ParseFloat(row[1], 64)

				if err != nil {
					log.Println("Error: Parse step_price:", err)
					continue
				}

				price.StepPrice = v

			case "provider":
				price.Provider = row[1]
			}
		}


	*/

	App.LockerPrices.Lock()

	// Проверяем есть ли мапа для этого провайдера
	if _, ok := App.Prices[price.Provider]; !ok {
		App.Prices[price.Provider] = make(map[string]*PriceTerminal)
	}

	old := App.Prices[price.Provider][price.SecCode]

	// Вкорячиваем
	if old == nil || old.DateTime < price.DateTime {
		App.Prices[price.Provider][price.SecCode] = &price
	}

	App.LockerPrices.Unlock()

	// Запустим в отдельном потоке чтоб ТИСИПИ не ждал математику эту
	go ExpressionExecutor(&price)

	//log.Println(">>>", price)
	//timeFinish := time.Now()
	//log.Printf("Логика отработала за %v нсек.", timeFinish.Sub(timeStart).Nanoseconds())

}

func onClose(client *TCP.Client) {
	log.Printf("Disconnect")
	Telega.Send(Configs.MyTelega, "Отключился какойто терминал: "+client.GetAddr())
}

func onConnect(client *TCP.Client) {
	log.Printf("Connect: %s", client.GetAddr())
	Telega.Send(Configs.MyTelega, "Подключился какойто терминал: "+client.GetAddr())
}

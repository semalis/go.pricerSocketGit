package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func showAll(w http.ResponseWriter, req *http.Request) {
	App.LockerPrices.RLock()
	str, err := json.Marshal(App.Prices)
	App.LockerPrices.RUnlock()

	if err != nil {
		log.Println("Error:", err.Error())
	}

	_, err = fmt.Fprint(w, string(str))

	if err != nil {
		log.Println("Error:", err.Error())
		return
	}
}

func showCode(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	//code := vars["code"]
	time_, err := strconv.ParseInt(vars["time"], 10, 64)

	if err != nil {
		time_ = 0
	}

	out := make(map[string]map[string]*PriceTerminal)

	App.LockerPrices.RLock()

	for _, v := range strings.Split(vars["code"], ",") {
		for prov, list := range App.Prices {
			if price := list[v]; price != nil {
				if price.DateTime > time_ {
					//out[prov] = make(map[string]*PriceTerminal)
					if _, ok := out[prov]; !ok {
						out[prov] = make(map[string]*PriceTerminal)
					}
					out[prov][price.SecCode] = price
				}
			}
		}
	}

	App.LockerPrices.RUnlock()

	str, err := json.Marshal(out)

	if err != nil {
		log.Println("Error:", err.Error())
	}

	_, err = fmt.Fprint(w, string(str))

	if err != nil {
		log.Println("Error:", err.Error())
		return
	}
}

func showProvider(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	//code := vars["code"]
	time_, err := strconv.ParseInt(vars["time"], 10, 64)

	if err != nil {
		time_ = 0
	}

	provider := vars["provider"]

	out := make(map[string]map[string]*PriceTerminal)

	App.LockerPrices.RLock()

	for _, v := range strings.Split(vars["code"], ",") {
		for prov, list := range App.Prices {
			if price := list[v]; price != nil {
				if price.DateTime > time_ && price.Provider == provider {
					if _, ok := out[prov]; !ok {
						out[prov] = make(map[string]*PriceTerminal)
					}
					out[prov][price.SecCode] = price
				}
			}
		}
	}

	App.LockerPrices.RUnlock()

	str, err := json.Marshal(out)

	if err != nil {
		log.Println("Error:", err.Error())
	}

	_, err = fmt.Fprint(w, string(str))

	if err != nil {
		log.Println("Error:", err.Error())
		return
	}
}

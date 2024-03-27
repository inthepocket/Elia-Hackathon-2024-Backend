package main

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
)

func calculateRoofPricePerQuarter(dayAheadPricesJson string) {
	headers := map[string]string{}
	headers["Content-Type"] = "application/json"
	params := url.Values{}

	dataJson := fmt.Sprintf("{\"time_series_data\": %s, \"ev_comfort_charge_capacity_kwh\": %d, \"ev_max_charge_capacity_kwh\": %d, \"buffer\": %f}",
		dayAheadPricesJson,
		50,
		100,
		1.0)

	log.Println("calculateRoofPricePerQuarter")
	log.Println(dataJson)

	body, err := makeRequest(os.Getenv("STEERING_PYTHON_URI"), "POST", "/calculate_roof_price_per_quarter", headers, params, bytes.NewBuffer([]byte(dataJson)))
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}

	log.Println(string(body))
}

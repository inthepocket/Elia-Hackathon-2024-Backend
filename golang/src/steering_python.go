package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
)

type RoofPrices struct {
	RoofComfort float32 `json:"roof_comfort"`
	RoofMax     float32 `json:"roof_max"`
}

func calculateRoofPricePerQuarter(dayAheadPricesJson string, evComfortChargeCapacityKwh int, evMaxChargeCapacityKwh int, bufferPerc float32) RoofPrices {
	log.Println("/// calculateRoofPricePerQuarter")
	headers := map[string]string{}
	headers["Content-Type"] = "application/json"
	params := url.Values{}

	dataJson := fmt.Sprintf("{\"time_series_data\": %s, \"ev_comfort_charge_capacity_kwh\": %d, \"ev_max_charge_capacity_kwh\": %d, \"buffer\": %f}",
		dayAheadPricesJson,
		evComfortChargeCapacityKwh,
		evMaxChargeCapacityKwh,
		bufferPerc)

	log.Println(dataJson)

	body, err := makeRequest("http://host.docker.internal:5001", "POST", "/calculate_roof_price_per_quarter", headers, params, bytes.NewBuffer([]byte(dataJson)))
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}
	log.Println(string(body))

	var roofPriceResponse RoofPrices
	if err := json.Unmarshal(body, &roofPriceResponse); err != nil {
		log.Println(err)
	}
	return roofPriceResponse

}

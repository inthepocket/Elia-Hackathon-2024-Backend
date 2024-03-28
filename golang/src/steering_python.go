package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
)

type RoofPrices struct {
	RoofComfort float32 `json:"roof_comfort"`
	RoofMax     float32 `json:"roof_max"`
	LastHourMax float32 `json:"last_hour_max"`
}

func calculateRoofPricePerQuarter(dayAheadPricesJson string, evComfortChargeCapacityKwh int, evMaxChargeCapacityKwh int, bufferPerc float32) (RoofPrices, error) {
	log.Println("/// calculateRoofPricePerQuarter", evComfortChargeCapacityKwh, evMaxChargeCapacityKwh, bufferPerc)
	headers := map[string]string{}
	headers["Content-Type"] = "application/json"
	params := url.Values{}

	dataJson := fmt.Sprintf("{\"time_series_data\": %s, \"ev_comfort_charge_capacity_kwh\": %d, \"ev_max_charge_capacity_kwh\": %d, \"buffer\": %f}",
		dayAheadPricesJson,
		evComfortChargeCapacityKwh,
		evMaxChargeCapacityKwh,
		bufferPerc)

	//log.Println(dataJson)

	body, err := makeRequest(os.Getenv("STEERING_PYTHON_URI"), "POST", "/calculate_roof_price_per_quarter", headers, params, bytes.NewBuffer([]byte(dataJson)))
	if err != nil {
		return RoofPrices{}, err
	}
	//log.Println(string(body))

	var roofPriceResponse RoofPrices
	if err := json.Unmarshal(body, &roofPriceResponse); err != nil {
		//log.Println("Error in steering-python", err)
		return RoofPrices{}, errors.New("Error in steering-python")
	}
	return roofPriceResponse, nil

}

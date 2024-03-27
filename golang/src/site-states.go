package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/url"
	"os"

	"github.com/tidwall/gjson"
)

type Car struct {
	Ean                            string
	Connected                      bool
	consumptionKwSincePreviousTime float32
}

func getActiveCars(token string) []Car {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/site-states/last", headers, params, nil)
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
		return []Car{}
	}
	log.Println(string(body))
	//result := gjson.Get(string(body), "assets.*.#($AssetType==\"EV\")#.ean")
	//log.Println(gjson.Get(string(body), "assets.*.ean"))
	//log.Println(gjson.Get(string(body), "assets.#(\\$AssetType==\"EV\")#.ean"))

	//gjson.Parse(string(body)).ForEach(func(key, value gjson.Result) bool {
	//	fmt.Println("Key:", key.String(), "Value:", value.String())
	//	return true // keep iterating
	//})

	var cars []Car
	gjson.GetBytes([]byte(string(body)), "assets").ForEach(func(key, value gjson.Result) bool {
		//if key.Raw == "$id" {
		//	return true
		//}
		//log.Println("Key:", key)
		//log.Println("Value:", value)
		//log.Println(gjson.Get(value.Raw, "$AssetType"))
		//log.Println(gjson.Get(value.Raw, "$AssetType").Str)
		if gjson.Get(value.Raw, "$AssetType").Str != "EV" {
			log.Println("not an EV")
			return true
		}

		var car Car
		car.Ean = gjson.Get(value.Raw, "ean").Str
		car.Connected = gjson.Get(value.Raw, "connected").Bool()
		if gjson.Get(value.Raw, "assetMode").Str == "Steered" {
			car.consumptionKwSincePreviousTime = float32(gjson.Get(value.Raw, "consumption").Num)
		}

		cars = append(cars, car)
		//log.Println(cars)

		return true // keep iterating
	})

	return cars
}

type InvidualSteeringRequest struct {
	Ean     string `json:"ean"`
	Steered bool   `json:"steered"`
	//RequestedProduction  float32 `json:"requestedProduction"`
	RequestedConsumption float32 `json:"requestedConsumption"`
}

type SteeringRequestData struct {
	RequestTime string                    `json:"requestTime"`
	Requests    []InvidualSteeringRequest `json:"requests"`
}

type InvidualProductionRequest struct {
	Ean                 string  `json:"ean"`
	Steered             bool    `json:"steered"`
	RequestedProduction float32 `json:"requestedProduction"`
	//RequestedConsumption float32 `json:"requestedConsumption"`
}

type ProductionRequestData struct {
	RequestTime string                    `json:"requestTime"`
	Requests    []InvidualSteeringRequest `json:"requests"`
}

func steeringRequest(token string, currentTime string, cars []Car, charge bool) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	headers["Content-Type"] = "application/json"

	params := url.Values{}

	dataRequest := SteeringRequestData{}
	dataRequest.RequestTime = roundToNext20Seconds(currentTime)
	for _, car := range cars {
		var request InvidualSteeringRequest
		request.Ean = car.Ean
		if car.Connected && charge {
			request.Steered = true
			request.RequestedConsumption = 22
		} else {
			request.Steered = true
			request.RequestedConsumption = 0
		}

		dataRequest.Requests = append(dataRequest.Requests, request)
	}
	//log.Println("cars", cars)
	//log.Println("dataRequest", dataRequest)
	jsonDataRequest, _ := json.Marshal([]SteeringRequestData{dataRequest})

	log.Println("######### REQUEST ############")
	log.Println(string(jsonDataRequest))
	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "POST", "/assets/steering-requests", headers, params, bytes.NewBuffer([]byte(jsonDataRequest)))
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
	}
	log.Println(string(body))
	log.Println("#####################")

}

func steeringRequestBattery(token string, currentTime string, cars []Car, charge bool) {

	var data []byte
	if charge {
		dataRequest := SteeringRequestData{}
		dataRequest.RequestTime = roundToNext20Seconds(currentTime)
		var request InvidualSteeringRequest
		request.Ean = "541657038024211911"
		request.Steered = true
		request.RequestedConsumption = 50
		data, _ = json.Marshal([]SteeringRequestData{dataRequest})
	} else {
		dataRequest := ProductionRequestData{}
		dataRequest.RequestTime = roundToNext20Seconds(currentTime)
		var request InvidualProductionRequest
		request.Ean = "541657038024211911"
		request.Steered = true
		request.RequestedProduction = 50
		data, _ = json.Marshal([]ProductionRequestData{dataRequest})
	}

	log.Println(data)

}

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
	Soc                            float32
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
	//log.Println(string(body))
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
			return true
		}

		var car Car
		car.Ean = gjson.Get(value.Raw, "ean").Str
		car.Connected = gjson.Get(value.Raw, "connected").Bool()
		if gjson.Get(value.Raw, "assetMode").Str == "Steered" {
			car.consumptionKwSincePreviousTime = float32(gjson.Get(value.Raw, "consumption").Num)
		}
		car.Soc = float32(gjson.Get(value.Raw, "soc").Num)

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
	RequestTime string                      `json:"requestTime"`
	Requests    []InvidualProductionRequest `json:"requests"`
}

func steeringRequest(token string, currentTime string, cars []Car, charge bool, carMinChargeLevels map[string]float32) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	headers["Content-Type"] = "application/json"

	params := url.Values{}

	dataRequest := SteeringRequestData{}
	dataRequest.RequestTime = roundToNext20Seconds(currentTime)
	for _, car := range cars {
		rechargeAnyway := false
		if carMinChargeLevel, ok := carMinChargeLevels[car.Ean]; ok {
			rechargeAnyway = carMinChargeLevel > car.Soc
			log.Println("#### Car", car.Ean, "recharging to minimum level of", car.Soc, " ####")
		}

		var request InvidualSteeringRequest
		request.Ean = car.Ean
		if (car.Connected && charge) || rechargeAnyway {
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

	// log.Println("######### REQUEST ############")
	// log.Println(string(jsonDataRequest))
	_, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "POST", "/assets/steering-requests", headers, params, bytes.NewBuffer([]byte(jsonDataRequest)))
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
	}
	// log.Println(string(body))
	// log.Println("#####################")

}

func steeringRequestBattery(token string, currentTime string, charge bool) {
	var data []byte
	if charge {
		dataRequest := SteeringRequestData{}
		dataRequest.RequestTime = roundToNext20Seconds(currentTime)
		var request InvidualSteeringRequest
		request.Ean = "541657038024211911"
		request.Steered = true
		request.RequestedConsumption = 50
		dataRequest.Requests = append(dataRequest.Requests, request)
		data, _ = json.Marshal([]SteeringRequestData{dataRequest})
	} else {
		dataRequest := ProductionRequestData{}
		dataRequest.RequestTime = roundToNext20Seconds(currentTime)
		var request InvidualProductionRequest
		request.Ean = "541657038024211911"
		request.Steered = true
		request.RequestedProduction = 50
		dataRequest.Requests = append(dataRequest.Requests, request)
		data, _ = json.Marshal([]ProductionRequestData{dataRequest})
	}

	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	headers["Content-Type"] = "application/json"

	params := url.Values{}

	// log.Println("######### REQUEST ############")
	// log.Println(string(data))
	_, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "POST", "/assets/steering-requests", headers, params, bytes.NewBuffer([]byte(data)))
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
	}
	// log.Println(string(body))
	// log.Println("#####################")

}

func steeringRequestSolar(token string, currentTime string, produce bool) {
	var data []byte
	if produce {
		dataRequest := ProductionRequestData{}
		dataRequest.RequestTime = roundToNext20Seconds(currentTime)
		var request InvidualProductionRequest
		request.Ean = "541787622019220646"
		request.Steered = true
		request.RequestedProduction = 200
		dataRequest.Requests = append(dataRequest.Requests, request)
		data, _ = json.Marshal([]ProductionRequestData{dataRequest})
	} else {
		dataRequest := ProductionRequestData{}
		dataRequest.RequestTime = roundToNext20Seconds(currentTime)
		var request InvidualProductionRequest
		request.Ean = "541787622019220646"
		request.Steered = true
		request.RequestedProduction = 0
		dataRequest.Requests = append(dataRequest.Requests, request)
		data, _ = json.Marshal([]ProductionRequestData{dataRequest})
	}

	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	headers["Content-Type"] = "application/json"

	params := url.Values{}

	// log.Println("######### REQUEST ############")
	// log.Println(string(data))
	_, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "POST", "/assets/steering-requests", headers, params, bytes.NewBuffer([]byte(data)))
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
	}
	// log.Println(string(body))
	// log.Println("#####################")

}

package main

import (
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/tidwall/gjson"
)

func getDayAheadPrices(token string, dateString string) string {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}
	params.Add("startDate", dateString+"T00:15:00+01:00")
	params.Add("endDate", getNextDay(dateString)+"T00:15:00+01:00")
	log.Println("/// getDayAheadPrices ", params)

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/prices/day-ahead-prices", headers, params, nil)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}

	return string(body)
}

func getLastPriceEstimate(priceEstimatesJson string) float64 {
	//log.Println("//////////////////")
	//log.Println(priceEstimatesJson)
	l := gjson.Get(priceEstimatesJson, "$values.0.priceEstimations.$values")
	//log.Println(l)
	item := l.Array()[len(l.Array())-1]
	//log.Println(item)
	//log.Println(item.Raw)
	price := gjson.Get(item.Raw, "price")
	//log.Println(price.Raw)
	floatPrice, _ := strconv.ParseFloat(price.Raw, 32)
	//log.Println(floatPrice)
	return floatPrice
}

func getRealTimePrice(token string, realTime string) float64 {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}
	params.Add("startDate", roundToPrevious15Mins(realTime))
	params.Add("endDate", roundToPrevious15Mins(realTime))
	log.Println("/// getRealTimePrice", params)

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/prices/full-realtime-prices", headers, params, nil)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}

	//log.Println(string(body))
	return getLastPriceEstimate(string(body))
}

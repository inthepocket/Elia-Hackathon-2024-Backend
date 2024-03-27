package main

import (
	"encoding/json"
	"errors"
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

func getLastPriceEstimate(priceEstimatesJson string) (float64, error) {
	l := gjson.Get(priceEstimatesJson, "$values.0.priceEstimations.$values")
	arr := l.Array()
	if len(arr) == 0 {
		return 0, errors.New("no price estimates available")
	}
	item := arr[len(arr)-1]
	price := gjson.Get(item.Raw, "price")
	floatPrice, err := strconv.ParseFloat(price.Raw, 32)
	if err != nil {
		return 0, err
	}
	return floatPrice, nil
}

func getRealTimePrice(token string, realTime string) (float64, error) {
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
	price, _ := getLastPriceEstimate(string(body))
	return price, nil
}

type RealTimePrice struct {
	date        string
	requestTime string
	price       float64
}

func getHistoricRealTimePrices(token, startDate, endDate string) []RealTimePrice {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}
	params.Add("startDate", roundToPrevious15Mins(startDate))
	params.Add("endDate", roundToPrevious15Mins(endDate))
	log.Println("/// getRealTimePrice", params)

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/prices/realtime-prices", headers, params, nil)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}

	var response []RealTimePrice
	if err := json.Unmarshal(body, &response); err != nil {
		log.Println(err)
		return []RealTimePrice{}
	}

	return response
}

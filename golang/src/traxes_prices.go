package main

import (
	"log"
	"net/url"
	"os"
)

func getDayAheadPrices(token string, dateString string) string {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}

	params.Add("startDate", dateString+"T00:15:00+01:00")
	params.Add("endDate", getNextDay(dateString)+"T00:15:00+01:00")
	log.Println("getDayAheadPrices ", params)

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/prices/day-ahead-prices", headers, params, nil)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}

	return string(body)

	/*var hackathonTimeResponse HackathonTimeResponse
	if err := json.Unmarshal(body, &hackathonTimeResponse); err != nil {
		log.Fatalln(err)
	}*/

}

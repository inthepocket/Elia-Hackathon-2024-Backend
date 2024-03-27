package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"
)

type HackathonTimeResponse struct {
	ID            string `json:"$id"`
	RequestTime   string `json:"requestTime"`
	HackathonTime string `json:"hackathonTime"`
}

func getCurrentHackathonTime(token string) string {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/times/HackathonTimeForNow", headers, nil, nil)
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
	}

	var hackathonTimeResponse HackathonTimeResponse
	if err := json.Unmarshal(body, &hackathonTimeResponse); err != nil {
		log.Println(err)
	}

	return hackathonTimeResponse.HackathonTime
}

func getDateString(hackathonTime string) string {
	return hackathonTime[0:10]
}

func getNextDay(dateString string) string {
	parsed, _ := time.Parse("2006-01-02", dateString)
	parsed = parsed.AddDate(0, 0, 1)
	return parsed.Format("2006-01-02")
}

func getHackathonTime(token, realTime string) (string, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}
	params.Add("realTime", realTime)

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/times/HackathonTimeForDateTime", headers, params, nil)
	if err != nil {
		return "", err
	}

	var hackathonTimeResponse HackathonTimeResponse
	if err := json.Unmarshal(body, &hackathonTimeResponse); err != nil {
		return "", err
	}

	return hackathonTimeResponse.HackathonTime, nil
}

package main

import (
	"encoding/json"
	"log"
	"net/url"
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

	body, err := makeRequest("GET", "/times/HackathonTimeForNow", headers, nil)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}

	var hackathonTimeResponse HackathonTimeResponse
	if err := json.Unmarshal(body, &hackathonTimeResponse); err != nil {
		log.Fatalln(err)
	}

	return hackathonTimeResponse.HackathonTime
}

func getHackathonTime(token, realTime string) string {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}
	params.Add("realTime", realTime)

	body, err := makeRequest("GET", "/times/HackathonTimeForDateTime", headers, params)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}

	var hackathonTimeResponse HackathonTimeResponse
	if err := json.Unmarshal(body, &hackathonTimeResponse); err != nil {
		log.Fatalln(err)
	}

	return hackathonTimeResponse.HackathonTime
}

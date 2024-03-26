package main

import (
	"encoding/json"
	"log"
)

type HackathonTimeResponse struct {
	ID            string `json:"$id"`
	RequestTime   string `json:"requestTime"`
	HackathonTime string `json:"hackathonTime"`
}

func getHackathonTime(token string) string {
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

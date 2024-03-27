package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"
)

type AssetState struct {
	ID                   string  `json:"$id"`
	AssetType            string  `json:"$AssetType"`
	SocMax               float32 `json:"socMax"`
	ChargingMax          float32 `json:"chargingMax"`
	Soc                  float32 `json:"soc"`
	LastSoc              float32 `json:"lastSoc"`
	Connected            bool    `json:"connected"`
	EmptyOnReconnect     bool    `json:"emptyOnReconnect"`
	StateTime            string  `json:"stateTime"`
	Ean                  string  `json:"ean"`
	AssetMode            string  `json:"assetMode"`
	MaxProduction        float32 `json:"maxProduction"`
	Production           float32 `json:"production"`
	RequestedProduction  float32 `json:"requestedProduction"`
	Consumption          float32 `json:"consumption"`
	RequestedConsumption float32 `json:"requestedConsumption"`
	SteerableConsumption bool    `json:"steerableConsumption"`
	SteerableProduction  bool    `json:"steerableProduction"`
}

type AssetStatesResponse struct {
	ID     string       `json:"$id"`
	Values []AssetState `json:"$values"`
}

func getHistoricAssetStates(token, ean, startDate, endDate string) []AssetState {

	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}
	params.Add("ean", ean)
	params.Add("startDate", startDate)
	params.Add("endDate", endDate)

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/assets/states", headers, params, nil)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}

	var assetStatesResponse AssetStatesResponse
	if err := json.Unmarshal(body, &assetStatesResponse); err != nil {
		log.Fatalln(err)
	}

	return assetStatesResponse.Values
}

type Session struct {
	StartState *AssetState `json:"startState"`
	EndState   *AssetState `json:"endState"`
}

func getAssetSessionsForDay(token, ean, realTime string) []Session {
	hackathonTime := getHackathonTime(token, realTime)

	var sessions []Session
	var startState, endState *AssetState

	// Parse the date string into a time.Time
	parsedDate, err := time.Parse(time.RFC3339, hackathonTime)
	if err != nil {
		log.Fatal("Error parsing date: ", err.Error())
	}

	// Start and end of the day
	startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location()).Format(time.RFC3339)
	endOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 0, parsedDate.Location()).Format(time.RFC3339)

	// Get the historic asset states for the specified date
	assetStates := getHistoricAssetStates(token, ean, startOfDay, endOfDay)

	for _, state := range assetStates {

		if err != nil {
			log.Fatal("Error parsing timestamp: ", err.Error())
		}

		if state.Connected && startState == nil {
			// Start of a new session
			startState = &state
		} else if !state.Connected && startState != nil {
			// End of the current session
			endState = &state
			sessions = append(sessions, Session{StartState: startState, EndState: endState})
			startState, endState = nil, nil
		}

	}

	// If there's an ongoing session at the end of the day, add it to the sessions
	if startState != nil {
		sessions = append(sessions, Session{StartState: startState, EndState: nil})
	}

	return sessions
}

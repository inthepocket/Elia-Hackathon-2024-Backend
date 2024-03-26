package main

import (
	"encoding/json"
	"log"
	"net/url"
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

	body, err := makeRequest("GET", "/assets/states", headers, params)
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

func getAssetSessionsForDay(token, ean, date string) []Session {
	var sessions []Session
	var startState, endState *AssetState

	// Parse the date string into a time.Time
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Fatal("Error parsing date: ", err.Error())
	}

	// Start and end of the day
	startOfDay := parsedDate.Format("2006-01-02T15:04:05Z")
	endOfDay := parsedDate.Add(24 * time.Hour).Format("2006-01-02T15:04:05Z")

	// Get the historic asset states for the specified date
	assetStates := getHistoricAssetStates(token, ean, startOfDay, endOfDay)

	for _, state := range assetStates {
		// Parse the state's timestamp
		stateTime, err := time.Parse(time.RFC3339, state.StateTime)
		if err != nil {
			log.Fatal("Error parsing timestamp: ", err.Error())
		}

		// Check if the state is within the day
		if stateTime.After(parsedDate) && stateTime.Before(parsedDate.Add(24*time.Hour)) {
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
	}

	// If there's an ongoing session at the end of the day, add it to the sessions
	if startState != nil {
		sessions = append(sessions, Session{StartState: startState, EndState: nil})
	}

	return sessions
}

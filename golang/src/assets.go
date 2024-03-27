package main

import (
	"encoding/json"
	"errors"
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

func getHistoricAssetStates(token, ean, startDate, endDate string) ([]AssetState, error) {

	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := url.Values{}
	params.Add("ean", ean)
	params.Add("startDate", startDate)
	params.Add("endDate", endDate)

	body, err := makeRequest(os.Getenv("TRAXES_API_BASE_URI"), "GET", "/assets/states", headers, params, nil)
	if err != nil {
		log.Println("Error on dispatching request. ", err.Error())
		return nil, err
	}

	var assetStatesResponse AssetStatesResponse
	if err := json.Unmarshal(body, &assetStatesResponse); err != nil {
		log.Println(err)
		return nil, err
	}

	if assetStatesResponse.Values == nil {
		log.Println("No asset states found for the specified date")
		return nil, errors.New("No asset states found for the specified date")
	}

	return assetStatesResponse.Values, nil
}

type Session struct {
	StartState   *AssetState `json:"startState"`
	ChargedState *AssetState `json:"chargedState"`
	EndState     *AssetState `json:"endState"`
}

func getAssetSessionsForDay(token, ean, realTime string) ([]Session, error) {
	hackathonTime, err := getHackathonTime(token, realTime)
	if err != nil {
		log.Println("Error getting hackathon time: ", err.Error())
		return nil, err
	}

	var sessions []Session
	var startState, chargedState, endState *AssetState

	// Parse the date string into a time.Time
	parsedDate, err := time.Parse(time.RFC3339, hackathonTime)
	if err != nil {
		log.Println("Error parsing date: ", err.Error())
		return nil, err
	}

	// Subtract 24 hours from the parsed date to get the start of the day
	startDate := parsedDate.Add(-24 * time.Hour).Format(time.RFC3339)

	// The end of the day is the parsed date
	endDate := parsedDate.Format(time.RFC3339)

	// Get the historic asset states for the specified date
	assetStates, err := getHistoricAssetStates(token, ean, startDate, endDate)
	if err != nil {
		log.Println("Error getting asset states: ", err.Error())
		return nil, err
	}

	for _, state := range assetStates {
		if state.Connected && startState == nil {
			// Start of a new session
			startState = &state
		} else if !state.Connected && startState != nil {
			// End of the current session
			endState = &state
			sessions = append(sessions, Session{StartState: startState, ChargedState: chargedState, EndState: endState})
			startState, chargedState, endState = nil, nil, nil
		} else if state.Soc == state.SocMax && state.Connected && chargedState == nil {
			// Fully charged state
			chargedState = &state
		}
	}

	// If there's an ongoing session at the end of the day, add it to the sessions
	if startState != nil {
		sessions = append(sessions, Session{StartState: startState, ChargedState: chargedState, EndState: nil})
	}

	return sessions, nil
}

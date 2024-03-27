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

type ChargePeriod struct {
	StartTime  string
	EndTime    string
	SocAtStart float32
	SocAtEnd   float32
	ChargedKwh float32
}

type Session struct {
	StartState    *AssetState `json:"StartState"`
	EndState      *AssetState `json:"EndState"`
	ChargePeriods []ChargePeriod
}

func getAssetSessionsForDay(token, ean, realTime string) ([]Session, error) {
	currentHackathonTime, err := getCurrentHackathonTime(token)
	if err != nil {
		log.Println("Error getting hackathon time: ", err.Error())
		return nil, err
	}

	hackathonTime, err := getHackathonTime(token, realTime)

	if err != nil {
		log.Println("Error getting hackathon time: ", err.Error())
		return nil, err
	}

	// Parse the date string into a time.Time
	parsedDate, err := time.Parse(time.RFC3339, hackathonTime)
	if err != nil {
		log.Println("Error parsing date: ", err.Error())
		return nil, err
	}

	// The start of the day is the parsed date with the time set to 00:00
	startDate := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location()).Format(time.RFC3339)

	// Parse the current hackathon time
	currentHackathonTimeParsed, err := time.Parse(time.RFC3339, currentHackathonTime)
	if err != nil {
		log.Println("Error parsing current hackathon time: ", err.Error())
		return nil, err
	}

	// The end of the day is the parsed date with the time set to 23:59, or the current hackathon time if it's in the same day
	endDate := parsedDate
	if parsedDate.Year() == currentHackathonTimeParsed.Year() && parsedDate.Month() == currentHackathonTimeParsed.Month() && parsedDate.Day() == currentHackathonTimeParsed.Day() {
		endDate = currentHackathonTimeParsed.Add(-5 * time.Minute)
	} else {
		endDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 0, parsedDate.Location())
	}
	endDateStr := endDate.Format(time.RFC3339)

	log.Println("Sessions for period: ", startDate, endDateStr)

	// Get the historic asset states for the specified date
	assetStates, err := getHistoricAssetStates(token, ean, startDate, endDateStr)
	if err != nil {
		log.Println("Error getting asset states: ", err.Error())
		return nil, err
	}

	log.Println("assetStates found: ", len(assetStates))

	//print amount of states with connected true
	connectedStates := 0
	for _, state := range assetStates {
		if state.Connected {
			connectedStates++
		}
	}
	log.Println("connectedStates: ", connectedStates)

	log.Println("DisconnectedStates: ", len(assetStates)-connectedStates)

	log.Println("\n\r")

	var prevState *AssetState

	var sessions []Session
	var currentSession *Session
	var currentChargePeriod *ChargePeriod

	for i := range assetStates {
		state := &assetStates[i]

		if prevState == nil {
			prevState = state
			continue
		}

		if state.Connected && !prevState.Connected {
			// Start of a new session
			currentSession = &Session{StartState: state}
		} else if !state.Connected && prevState.Connected && currentSession != nil {
			// End of the current session
			currentSession.EndState = state
			sessions = append(sessions, *currentSession)
			currentSession = nil
		}

		if state.Consumption > 0 && (currentChargePeriod == nil || currentChargePeriod.EndTime != "") {
			// Start of a new charge period
			currentChargePeriod = &ChargePeriod{StartTime: state.StateTime, SocAtStart: state.Soc}
		} else if state.Consumption == 0 && currentChargePeriod != nil && currentChargePeriod.EndTime == "" {
			// End of the current charge period
			currentChargePeriod.EndTime = state.StateTime
			currentChargePeriod.SocAtEnd = state.Soc
			currentChargePeriod.ChargedKwh = currentChargePeriod.SocAtEnd - currentChargePeriod.SocAtStart
			if currentSession != nil && currentChargePeriod != nil {
				currentSession.ChargePeriods = append(currentSession.ChargePeriods, *currentChargePeriod)
			}
		}

		prevState = state
	}

	// If there's an ongoing session or charge period when the asset states end, add them to the sessions
	if currentSession != nil {
		if currentChargePeriod != nil && currentChargePeriod.EndTime == "" {
			currentChargePeriod.EndTime = assetStates[len(assetStates)-1].StateTime
			currentChargePeriod.SocAtEnd = assetStates[len(assetStates)-1].Soc
			currentChargePeriod.ChargedKwh = currentChargePeriod.SocAtEnd - currentChargePeriod.SocAtStart
			if currentSession != nil && currentChargePeriod != nil {
				currentSession.ChargePeriods = append(currentSession.ChargePeriods, *currentChargePeriod)
			}
		}
		sessions = append(sessions, *currentSession)
	}

	return sessions, nil
}

func getCurrentAssetState(token, ean string) (*AssetState, error) {
	hackathonTime, err := getCurrentHackathonTime(token)
	if err != nil {
		log.Println("Error getting hackathon time: ", err.Error())
		return nil, err
	}

	parsedDate, err := time.Parse(time.RFC3339, hackathonTime)
	if err != nil {
		log.Println("Error parsing date: ", err.Error())
		return nil, err
	}

	startDate := parsedDate.Add(-40 * time.Second).Format(time.RFC3339)
	endDate := parsedDate.Add(-20 * time.Second).Format(time.RFC3339)

	log.Println("Current time: ", startDate, endDate)

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

	log.Println("Asset state response: ", string(body))

	type AssetStateApiResponse struct {
		ID     string       `json:"$id"`
		Values []AssetState `json:"$values"`
	}

	var response AssetStateApiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Println(err)
		return nil, err
	}

	if len(response.Values) == 0 {
		return nil, errors.New("no asset state returned by the API")
	}

	return &response.Values[0], nil

}

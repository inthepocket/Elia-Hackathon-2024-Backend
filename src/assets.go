package main

import (
	"encoding/json"
	"log"
	"net/url"
)

type AssetState struct {
	ID                   string `json:"$id"`
	AssetType            string `json:"$AssetType"`
	SocMax               int    `json:"socMax"`
	ChargingMax          int    `json:"chargingMax"`
	Soc                  int    `json:"soc"`
	LastSoc              int    `json:"lastSoc"`
	Connected            bool   `json:"connected"`
	EmptyOnReconnect     bool   `json:"emptyOnReconnect"`
	StateTime            string `json:"stateTime"`
	Ean                  string `json:"ean"`
	AssetMode            string `json:"assetMode"`
	MaxProduction        int    `json:"maxProduction"`
	Production           int    `json:"production"`
	RequestedProduction  int    `json:"requestedProduction"`
	Consumption          int    `json:"consumption"`
	RequestedConsumption int    `json:"requestedConsumption"`
	SteerableConsumption bool   `json:"steerableConsumption"`
	SteerableProduction  bool   `json:"steerableProduction"`
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

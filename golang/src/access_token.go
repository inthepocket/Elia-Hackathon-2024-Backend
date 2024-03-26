package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

func GetAccessToken() string {

	authUri := os.Getenv("TRAXES_AUTH_URI")
	clientID := os.Getenv("TRAXES_API_CLIENT_ID")
	clientSecret := os.Getenv("TRAXES_API_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" || authUri == "" {
		log.Fatalln("Environment variables TRAXES_CLIENT_ID and TRAXES_CLIENT_SECRET must be set")
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", "esp")

	req, err := http.NewRequest("POST", authUri, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal("Error on creating request object. ", err.Error())
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(body))

	// Unmarshal the JSON response into the TokenResponse struct
	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		log.Fatalln(err)
	}

	return tokenResponse.AccessToken
}

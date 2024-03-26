package access_token

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

func GetAccessToken() string {

	var authUrl = os.Getenv("TRAXES_AUTH_URI")

	// Create a URL
	u, err := url.Parse(authUrl)
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("grant_type", "client_credentials")
	params.Add("client_id", os.Getenv("TRAXES_API_CLIENT_ID"))
	params.Add("client_secret", os.Getenv("TRAXES_API_CLIENT_SECRET"))

	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// Unmarshal the JSON response into the TokenResponse struct
	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		log.Fatalln(err)
	}

	return tokenResponse.AccessToken
}

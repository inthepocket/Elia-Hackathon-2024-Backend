package main

import (
	"io"
	"net/http"
	"net/url"
)

func makeRequest(baseUrl string, method, path string, headers map[string]string, params url.Values, data io.Reader) ([]byte, error) {
	u, err := url.Parse(baseUrl + path)
	if err != nil {
		return nil, err
	}

	u.RawQuery = params.Encode()

	req, err := http.NewRequest(method, u.String(), data)
	if err != nil {
		return nil, err
	}

	//log.Println("Requesting: ", u.String())

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

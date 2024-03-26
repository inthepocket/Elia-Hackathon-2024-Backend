package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
)

func makeRequest(method, path string, headers map[string]string, params url.Values) ([]byte, error) {
	u, err := url.Parse(os.Getenv("TRAXES_API_BASE_URI") + path)
	if err != nil {
		return nil, err
	}

	u.RawQuery = params.Encode()

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

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

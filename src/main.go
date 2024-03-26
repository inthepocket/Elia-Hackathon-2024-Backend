package main

import (
	"errors"
	"log"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, _ *http.Request) {
	log.Println("received GET /hello")

	w.Write([]byte("Hello, world!"))

}

func main() {
	accessToken := GetAccessToken()
	assetStates := getHistoricAssetStates(accessToken, "541787622019220646", "2024-01-01T15:00:00Z", "2024-01-01T16:00:00Z")

	log.Println(assetStates)

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handlerFunc)

	server := http.Server{
		Addr:    ":80",
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Error running http server: %s\n", err)
		}
	}

}

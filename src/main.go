package main

import (
	"encoding/json"
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

	sessionsToday := getAssetSessionsForDay(accessToken, "541983310278725782", "2024-01-10")

	log.Println(sessionsToday)

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handlerFunc)

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		log.Println("received GET /sessions")

		// Parse the query parameters
		query := r.URL.Query()
		ean := query.Get("ean")
		date := query.Get("date")

		// Get the asset sessions for the specified day
		sessions := getAssetSessionsForDay(accessToken, ean, date)

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(sessions); err != nil {
			log.Printf("Error encoding response: %s\n", err)
		}
	})

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

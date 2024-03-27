package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

func handlerFunc(w http.ResponseWriter, _ *http.Request) {
	log.Println("received GET /hello")

	w.Write([]byte("Hello, world!"))

}

func main() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)

	if err := godotenv.Load(exPath + "/.env"); err != nil {
		log.Println("No .env file found")
	}

	mongo := getMongoClient()

	vehicles := getAllVehicles(mongo)

	log.Println("Vehicles:", vehicles)

	accessToken := GetAccessToken()

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handlerFunc)

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		log.Println("received GET /sessions")

		// Parse the query parameters
		query := r.URL.Query()
		ean := query.Get("ean")
		realTime := query.Get("realTime")
		realTime, err := url.QueryUnescape(realTime)
		if err != nil {
			log.Fatal(err)
		}

		// Get the asset sessions for the specified day
		sessions := getAssetSessionsForDay(accessToken, ean, realTime)

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(sessions); err != nil {
			log.Printf("Error encoding response: %s\n", err)
		}
	})

	time.Sleep(time.Second * 5)
	go steerAssets(accessToken)

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

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

	log.Println("Vehicles:", vehicles[0])

	accessToken := GetAccessToken()

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handlerFunc)

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		log.Println("received GET /sessions", r)

		// Parse the query parameters
		query := r.URL.Query()

		ean := query.Get("ean")
		realTime := query.Get("realTime")

		if ean == "" {
			http.Error(w, "Missing required 'ean' query parameter", http.StatusBadRequest)
			return
		}

		if realTime == "" {
			http.Error(w, "Missing required 'realTime' query parameter", http.StatusBadRequest)
			return
		}

		realTime, err := url.QueryUnescape(realTime)
		if err != nil {
			log.Println(err)
		}

		// Get the asset sessions for the specified day
		sessions, err := getAssetSessionsForDay(accessToken, ean, realTime)
		if err != nil {
			http.Error(w, "Error getting asset sessions", http.StatusInternalServerError)
			return
		}

		if sessions == nil {
			sessions = []Session{}
		}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(sessions); err != nil {
			log.Printf("Error encoding response: %s\n", err)
		}
	})

	mux.HandleFunc("/vehicles", func(w http.ResponseWriter, r *http.Request) {
		log.Println("received GET /vehicles", r)

		// Parse the query parameters
		query := r.URL.Query()

		ean := query.Get("ean")
		if ean == "" {
			http.Error(w, "Missing required 'ean' query parameter", http.StatusBadRequest)
			return
		}

		vehicle := getVehicleByEan(mongo, ean)

		assetState, err := getCurrentAssetState(accessToken, ean)

		if err != nil {
			http.Error(w, "Error getting site state", http.StatusInternalServerError)
			return
		}

		log.Println("Asset state:", assetState)

		assetSessionsLast24h, _ := getAssetSessionsForDay(accessToken, ean, time.Now().Format(time.RFC3339))
		// if err != nil {
		// 	assetSessionsLast24h = []Session{}
		// 	return
		// }

		vehicleResponse := VehicleResponse{
			Metadata:            vehicle,
			CurrentState:        *assetState,
			SessionsLast24hours: assetSessionsLast24h,
		}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(vehicleResponse); err != nil {
			log.Printf("Error encoding response: %s\n", err)
		}

	})

	// time.Sleep(time.Second * 5)
	// go steerAssets(accessToken)

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

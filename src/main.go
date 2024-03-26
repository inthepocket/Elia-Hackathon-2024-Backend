package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/inthepocket/Elia-Hackathon-2024-Backend/internal/access_token"
)

func handlerFunc(w http.ResponseWriter, _ *http.Request) {
	log.Println("received GET /hello")

	var access_token = access_token.GetAccessToken()

	w.Write([]byte(access_token))

}

func main() {

	log.Println("Hello world")

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

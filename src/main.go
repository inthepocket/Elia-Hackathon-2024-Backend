package main

import (
	"errors"
	"log"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, _ *http.Request) {
	log.Println("received GET /hello")

	w.Write([]byte("Hello world from server"))

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

package main

import (
	"log"
	"net/http"
	"time"

	"ghappworkshop/internal/githubapp"
)

func main() {
	cfg, err := githubapp.ConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	app, err := githubapp.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           app.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("open http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

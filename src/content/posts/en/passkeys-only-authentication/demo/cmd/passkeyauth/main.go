package main

import (
	"log"
	"net/http"
	"time"

	"passkeyauth/internal/auth"
)

func main() {
	app, err := auth.New(auth.Config{
		RPDisplayName: "Passkey Workshop",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
		StaticDir:     "static",
	})
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

	log.Println("open http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

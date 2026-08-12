package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mnikitenkov/appchain-control-plane-demo/internal/configstore"
)

func main() {
	address := os.Getenv("HTTP_ADDRESS")
	if address == "" {
		address = ":8080"
	}
	server := &http.Server{
		Addr:              address,
		Handler:           configstore.NewHandler(configstore.NewService()),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("appchain control plane listening on %s", address)
	log.Fatal(server.ListenAndServe())
}


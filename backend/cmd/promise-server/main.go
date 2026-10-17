package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/promisegrid/promise-app/backend/internal/api"
	"github.com/promisegrid/promise-app/backend/internal/store"
)

func main() {
	dataDirectory := environment("PROMISE_DATA_DIR", "./data")
	address := environment("PROMISE_LISTEN_ADDR", "127.0.0.1:8080")
	data, err := store.Open(dataDirectory)
	if err != nil {
		log.Fatalf("open data store: %v", err)
	}
	server := &http.Server{
		Addr: address, Handler: api.New(data, os.Getenv("PROMISE_ADMIN_TOKEN"), slog.Default()).Handler(),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
	}
	log.Printf("promise relay listening on %s with data in %s", address, dataDirectory)
	log.Fatal(server.ListenAndServe())
}

func environment(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

package main

import (
	"log"
	"net/http"

	"github.com/watchroom/backend/internal/config"
	apphttp "github.com/watchroom/backend/internal/http"
	"github.com/watchroom/backend/internal/room"
	"github.com/watchroom/backend/internal/ws"
)

func main() {
	cfg := config.Load()

	manager := room.NewManager()
	hub := ws.NewHub()

	router := apphttp.NewRouter(manager, hub, cfg.AllowedOrigin)

	log.Printf("starting server on :%s (CORS: %s)", cfg.Port, cfg.AllowedOrigin)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"
	"net/http"

	"goal-manager/backend/internal/app"
)

func main() {
	cfg := app.LoadConfig()
	server := app.NewServer(cfg)

	log.Printf("goal-manager backend listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}

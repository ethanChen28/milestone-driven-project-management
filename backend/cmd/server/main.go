package main

import (
	"log"
	"net/http"

	"goal-manager/backend/internal/app"
)

func main() {
	cfg := app.LoadConfig()
	server, err := app.NewServerE(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("goal-manager backend listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}

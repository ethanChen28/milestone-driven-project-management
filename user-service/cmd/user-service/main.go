package main

import (
	"context"
	"log"
	"net/http"

	"goal-manager/user-service/internal/identity"
)

func main() {
	cfg := identity.LoadConfig()
	server, err := identity.NewServer(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("user-service listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}

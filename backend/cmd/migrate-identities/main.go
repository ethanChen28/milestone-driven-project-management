package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"goal-manager/backend/internal/service"
)

func main() {
	dsn := flag.String("dsn", os.Getenv("MYSQL_DSN"), "MySQL DSN containing app_state JSON")
	flag.Parse()
	if *dsn == "" {
		log.Fatal("MYSQL_DSN or -dsn is required")
	}
	repo, err := service.NewMySQLRepository(context.Background(), *dsn)
	if err != nil {
		log.Fatal(err)
	}
	store, err := service.NewStoreWithRepository(repo)
	if err != nil {
		log.Fatal(err)
	}
	report := store.IdentityMigrationReport(service.DefaultUserProfiles())
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(payload))
	if report.UnresolvedReferences > 0 {
		os.Exit(2)
	}
}

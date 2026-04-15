package main

import (
	"log"

	"cityhawk/backend/internal/app"
	appconfig "cityhawk/backend/internal/config"
)

func main() {
	cfg := appconfig.LoadFromEnv()
	server, cleanup, err := app.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	log.Printf("server started on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"

	"cityhawk/backend/internal/app"
	appconfig "cityhawk/backend/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, use system env")
	}

	cfg := appconfig.LoadFromEnv()
	server := app.NewServer(cfg)

	log.Printf("server started on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

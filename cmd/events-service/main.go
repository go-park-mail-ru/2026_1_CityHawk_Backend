package main

import (
	"log"
	"net"
	"os"

	"cityhawk/backend/internal/app"
	appconfig "cityhawk/backend/internal/config"
)

func main() {
	cfg := appconfig.LoadFromEnv()
	server, cleanup, err := app.NewEventsGRPCServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	addr := env("EVENTS_GRPC_ADDR", ":50053")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("events grpc service started on %s", addr)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

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
	server, cleanup, err := app.NewAuthGRPCServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	addr := env("AUTH_GRPC_ADDR", ":50051")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("auth grpc service started on %s", addr)
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

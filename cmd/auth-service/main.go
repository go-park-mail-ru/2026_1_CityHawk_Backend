package main

import (
	"log"
	"net"

	"cityhawk/backend/internal/app"
	appconfig "cityhawk/backend/internal/config"
	"cityhawk/backend/internal/platform/metricsserver"
)

func main() {
	cfg := appconfig.LoadFromEnv()
	server, cleanup, err := app.NewAuthGRPCServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	metricsCleanup, err := metricsserver.Start(cfg.Metrics.AuthAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer metricsCleanup()

	addr := cfg.GRPC.AuthAddr
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("auth grpc service started on %s", addr)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

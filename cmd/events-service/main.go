package main

import (
	"log"
	"net"

	"cityhawk/backend/internal/app"
	appconfig "cityhawk/backend/internal/config"
	platformmetrics "cityhawk/backend/internal/platform/metrics"
)

func main() {
	cfg := appconfig.LoadFromEnv()
	server, cleanup, err := app.NewEventsGRPCServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	metricsCleanup, err := platformmetrics.StartServer(cfg.Metrics.EventsAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer metricsCleanup()

	addr := cfg.GRPC.EventsAddr
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("events grpc service started on %s", addr)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

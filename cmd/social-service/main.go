package main

import (
	"log"
	"net"

	"cityhawk/backend/internal/app"
	appconfig "cityhawk/backend/internal/config"
)

func main() {
	cfg := appconfig.LoadFromEnv()
	server, cleanup, err := app.NewSocialGRPCServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	addr := cfg.GRPC.SocialAddr
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("social grpc service started on %s", addr)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

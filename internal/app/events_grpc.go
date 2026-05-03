package app

import (
	appconfig "cityhawk/backend/internal/config"
	photonintegration "cityhawk/backend/internal/integration/photon"
	placegrpc "cityhawk/backend/internal/place/delivery/grpc"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
	platformmetrics "cityhawk/backend/internal/platform/metrics"
	eventsv1 "cityhawk/backend/pkg/pb/events/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewEventsGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	pool, err := newGRPCPostgresPool(cfg)
	if err != nil {
		return nil, nil, err
	}

	repo := placerepo.NewPostgresRepository(pool)
	eventsUC := placeusecase.NewService(repo)

	var lookupUC *placeusecase.PlaceLookupService
	if cfg.Photon.Enabled {
		photonClient := photonintegration.NewClient(photonintegration.ClientConfig{
			BaseURL:        cfg.Photon.BaseURL,
			RequestTimeout: cfg.Photon.RequestTimeout,
		})
		lookupUC = placeusecase.NewPlaceLookupService(
			photonClient,
			repo,
			[]byte(cfg.Auth.JWTSecret),
			cfg.Photon.DefaultCountry,
			cfg.Photon.DefaultTimezone,
		)
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(platformmetrics.UnaryServerInterceptor("cityhawk-events-service")))
	eventsv1.RegisterEventsServiceServer(server, placegrpc.NewServer(eventsUC, lookupUC))
	reflection.Register(server)
	return server, pool.Close, nil
}

package app

import (
	appconfig "cityhawk/backend/internal/config"
	platformmetrics "cityhawk/backend/internal/platform/metrics"
	socialgrpc "cityhawk/backend/internal/social/delivery/grpc"
	socialrepo "cityhawk/backend/internal/social/repository"
	socialv1 "cityhawk/backend/pkg/pb/social/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewSocialGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	pool, err := newGRPCPostgresPool(cfg)
	if err != nil {
		return nil, nil, err
	}

	repo := socialrepo.NewPostgresRepository(pool)
	server := grpc.NewServer(grpc.UnaryInterceptor(platformmetrics.UnaryServerInterceptor("cityhawk-social-service")))
	socialv1.RegisterSocialServiceServer(server, socialgrpc.NewServer(repo))
	reflection.Register(server)
	return server, pool.Close, nil
}

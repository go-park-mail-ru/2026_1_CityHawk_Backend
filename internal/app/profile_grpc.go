package app

import (
	appconfig "cityhawk/backend/internal/config"
	platformmetrics "cityhawk/backend/internal/platform/metrics"
	usergrpc "cityhawk/backend/internal/user/delivery/grpc"
	userrepo "cityhawk/backend/internal/user/repository"
	profilev1 "cityhawk/backend/pkg/pb/profile/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewProfileGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	pool, err := newGRPCPostgresPool(cfg)
	if err != nil {
		return nil, nil, err
	}

	users := userrepo.NewPostgresUserRepository(pool)
	server := grpc.NewServer(grpc.UnaryInterceptor(platformmetrics.UnaryServerInterceptor("cityhawk-profile-service")))
	profilev1.RegisterProfileServiceServer(server, usergrpc.NewServer(users))
	reflection.Register(server)
	return server, pool.Close, nil
}

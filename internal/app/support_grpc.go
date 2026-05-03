package app

import (
	appconfig "cityhawk/backend/internal/config"
	platformmetrics "cityhawk/backend/internal/platform/metrics"
	supportgrpc "cityhawk/backend/internal/support/delivery/grpc"
	supportrepo "cityhawk/backend/internal/support/repository"
	supportusecase "cityhawk/backend/internal/support/usecase"
	profilegateway "cityhawk/backend/internal/user/gateway/profilegrpc"
	profilev1 "cityhawk/backend/pkg/pb/profile/v1"
	supportv1 "cityhawk/backend/pkg/pb/support/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func NewSupportGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	pool, err := newGRPCPostgresPool(cfg)
	if err != nil {
		return nil, nil, err
	}

	profileConn, err := grpc.NewClient(cfg.GRPC.ProfileAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	users := profilegateway.NewClient(profilev1.NewProfileServiceClient(profileConn))
	supportRepo := supportrepo.NewPostgresRepository(pool)
	supportUC := supportusecase.NewService(supportRepo, users)
	server := grpc.NewServer(grpc.UnaryInterceptor(platformmetrics.UnaryServerInterceptor("cityhawk-support-service")))
	supportv1.RegisterSupportServiceServer(server, supportgrpc.NewServer(supportUC))
	reflection.Register(server)
	cleanup := func() {
		_ = profileConn.Close()
		pool.Close()
	}
	return server, cleanup, nil
}

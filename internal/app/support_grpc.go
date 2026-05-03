package app

import (
	appconfig "cityhawk/backend/internal/config"
	supportgrpc "cityhawk/backend/internal/support/delivery/grpc"
	supportrepo "cityhawk/backend/internal/support/repository"
	supportusecase "cityhawk/backend/internal/support/usecase"
	userrepo "cityhawk/backend/internal/user/repository"
	supportv1 "cityhawk/backend/pkg/pb/support/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewSupportGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	pool, err := newGRPCPostgresPool(cfg)
	if err != nil {
		return nil, nil, err
	}

	users := userrepo.NewPostgresUserRepository(pool)
	supportRepo := supportrepo.NewPostgresRepository(pool)
	supportUC := supportusecase.NewService(supportRepo, users)
	server := grpc.NewServer()
	supportv1.RegisterSupportServiceServer(server, supportgrpc.NewServer(supportUC))
	reflection.Register(server)
	return server, pool.Close, nil
}

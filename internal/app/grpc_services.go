package app

import (
	"context"
	"log"

	authgrpc "cityhawk/backend/internal/auth/delivery/grpc"
	gatewaygoogle "cityhawk/backend/internal/auth/gateway/google"
	gatewayvk "cityhawk/backend/internal/auth/gateway/vk"
	gatewayyandex "cityhawk/backend/internal/auth/gateway/yandex"
	authrepo "cityhawk/backend/internal/auth/repository"
	authusecase "cityhawk/backend/internal/auth/usecase"
	appconfig "cityhawk/backend/internal/config"
	photonintegration "cityhawk/backend/internal/integration/photon"
	placegrpc "cityhawk/backend/internal/place/delivery/grpc"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
	platformpostgres "cityhawk/backend/internal/platform/postgres"
	platformsecurity "cityhawk/backend/internal/platform/security"
	socialgrpc "cityhawk/backend/internal/social/delivery/grpc"
	socialrepo "cityhawk/backend/internal/social/repository"
	supportgrpc "cityhawk/backend/internal/support/delivery/grpc"
	supportrepo "cityhawk/backend/internal/support/repository"
	supportusecase "cityhawk/backend/internal/support/usecase"
	usergrpc "cityhawk/backend/internal/user/delivery/grpc"
	userrepo "cityhawk/backend/internal/user/repository"
	authv1 "cityhawk/backend/pkg/pb/auth/v1"
	eventsv1 "cityhawk/backend/pkg/pb/events/v1"
	profilev1 "cityhawk/backend/pkg/pb/profile/v1"
	socialv1 "cityhawk/backend/pkg/pb/social/v1"
	supportv1 "cityhawk/backend/pkg/pb/support/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewAuthGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	ctx := context.Background()
	pool, err := platformpostgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		return nil, nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}

	users := userrepo.NewPostgresUserRepository(pool)
	refreshRepo := authrepo.NewPostgresRefreshRepository(pool)
	tokenService := platformsecurity.NewJWTTokenService([]byte(cfg.Auth.JWTSecret))
	tokenUC := authusecase.NewService(cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL, users, refreshRepo, tokenService)
	flowUC := authusecase.NewAuthFlowService(users, tokenUC, platformsecurity.NewBcryptPasswordService())
	oauthUsers := authusecase.NewOAuthUserService(users, platformsecurity.NewBcryptPasswordService())

	vkOAuthCfg, err := gatewayvk.NewOAuthConfig(cfg.OAuth.VK)
	if err != nil {
		log.Printf("vk oauth disabled: %v", err)
	}
	yandexOAuthCfg, err := gatewayyandex.NewOAuthConfig(cfg.OAuth.Yandex)
	if err != nil {
		log.Printf("yandex oauth disabled: %v", err)
	}
	googleOAuthCfg, err := gatewaygoogle.NewOAuthConfig(cfg.OAuth.Google)
	if err != nil {
		log.Printf("google oauth disabled: %v", err)
	}
	oauthUC := authusecase.NewOAuthLoginService(
		gatewaygoogle.NewGateway(googleOAuthCfg),
		gatewayyandex.NewGateway(yandexOAuthCfg),
		gatewayvk.NewGateway(vkOAuthCfg),
		oauthUsers,
		tokenUC,
	)

	server := grpc.NewServer()
	authv1.RegisterAuthServiceServer(server, authgrpc.NewServer(flowUC, tokenUC, oauthUC, users, tokenService))
	reflection.Register(server)
	return server, pool.Close, nil
}

func NewProfileGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	ctx := context.Background()
	pool, err := platformpostgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		return nil, nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}

	users := userrepo.NewPostgresUserRepository(pool)
	server := grpc.NewServer()
	profilev1.RegisterProfileServiceServer(server, usergrpc.NewServer(users))
	reflection.Register(server)
	return server, pool.Close, nil
}

func NewEventsGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	ctx := context.Background()
	pool, err := platformpostgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		return nil, nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
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

	server := grpc.NewServer()
	eventsv1.RegisterEventsServiceServer(server, placegrpc.NewServer(eventsUC, lookupUC))
	reflection.Register(server)
	return server, pool.Close, nil
}

func NewSupportGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	ctx := context.Background()
	pool, err := platformpostgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		return nil, nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
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

func NewSocialGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	ctx := context.Background()
	pool, err := platformpostgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		return nil, nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}

	repo := socialrepo.NewPostgresRepository(pool)
	server := grpc.NewServer()
	socialv1.RegisterSocialServiceServer(server, socialgrpc.NewServer(repo))
	reflection.Register(server)
	return server, pool.Close, nil
}

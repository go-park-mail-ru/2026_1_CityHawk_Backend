package app

import (
	"log"

	authgrpc "cityhawk/backend/internal/auth/delivery/grpc"
	gatewaygoogle "cityhawk/backend/internal/auth/gateway/google"
	gatewayvk "cityhawk/backend/internal/auth/gateway/vk"
	gatewayyandex "cityhawk/backend/internal/auth/gateway/yandex"
	authrepo "cityhawk/backend/internal/auth/repository"
	authusecase "cityhawk/backend/internal/auth/usecase"
	appconfig "cityhawk/backend/internal/config"
	platformmetrics "cityhawk/backend/internal/platform/metrics"
	platformsecurity "cityhawk/backend/internal/platform/security"
	userrepo "cityhawk/backend/internal/user/repository"
	authv1 "cityhawk/backend/pkg/pb/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewAuthGRPCServer(cfg appconfig.Config) (*grpc.Server, func(), error) {
	pool, err := newGRPCPostgresPool(cfg)
	if err != nil {
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

	server := grpc.NewServer(grpc.UnaryInterceptor(platformmetrics.UnaryServerInterceptor("cityhawk-auth-service")))
	authv1.RegisterAuthServiceServer(server, authgrpc.NewServer(flowUC, tokenUC, oauthUC, users, tokenService))
	reflection.Register(server)
	return server, pool.Close, nil
}

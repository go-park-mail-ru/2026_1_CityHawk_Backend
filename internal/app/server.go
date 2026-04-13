package app

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	authdelivery "cityhawk/backend/internal/auth/delivery/http"
	gatewaygoogle "cityhawk/backend/internal/auth/gateway/google"
	gatewayvk "cityhawk/backend/internal/auth/gateway/vk"
	gatewayyandex "cityhawk/backend/internal/auth/gateway/yandex"
	authrepo "cityhawk/backend/internal/auth/repository"
	authusecase "cityhawk/backend/internal/auth/usecase"
	appconfig "cityhawk/backend/internal/config"
	kudagointegration "cityhawk/backend/internal/integration/kudago"
	placedelivery "cityhawk/backend/internal/place/delivery/http"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	platformpostgres "cityhawk/backend/internal/platform/postgres"
	platformsecurity "cityhawk/backend/internal/platform/security"
	userdelivery "cityhawk/backend/internal/user/delivery/http"
	userrepo "cityhawk/backend/internal/user/repository"
)

func NewServer(cfg appconfig.Config) (*http.Server, func(), error) {
	ctx := context.Background()
	pool, err := platformpostgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		return nil, nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}
	syncCtx, syncCancel := context.WithCancel(context.Background())
	var syncWG sync.WaitGroup

	store := userrepo.NewPostgresUserRepository(pool)
	placeRepo := placerepo.NewPostgresRepository(pool)
	placeUC := placeusecase.NewService(placeRepo)
	refreshRepo := authrepo.NewPostgresRefreshRepository(pool)
	tokenService := platformsecurity.NewJWTTokenService([]byte(cfg.Auth.JWTSecret))
	authUC := authusecase.NewService(cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL, store, refreshRepo, tokenService)
	authFlowUC := authusecase.NewAuthFlowService(
		store,
		authUC,
		platformsecurity.NewBcryptPasswordService(),
	)
	oauthUsers := authusecase.NewOAuthUserService(
		store,
		platformsecurity.NewBcryptPasswordService(),
	)
	placeHandler := placedelivery.NewHandler(placeUC)
	authFlowHandler := authdelivery.NewAuthHandler(authFlowUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	authRefreshHandler := authdelivery.NewRefreshHandler(authUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	meHandler := userdelivery.NewMeHandler(store)
	vkOAuthCfg, _ := gatewayvk.NewOAuthConfig(cfg.OAuth.VK)
	yandexOAuthCfg, _ := gatewayyandex.NewOAuthConfig(cfg.OAuth.Yandex)
	googleOAuthCfg, _ := gatewaygoogle.NewOAuthConfig(cfg.OAuth.Google)
	googleGateway := gatewaygoogle.NewGateway(googleOAuthCfg)
	yandexGateway := gatewayyandex.NewGateway(yandexOAuthCfg)
	vkGateway := gatewayvk.NewGateway(vkOAuthCfg)
	oauthLoginUC := authusecase.NewOAuthLoginService(googleGateway, yandexGateway, vkGateway, oauthUsers, authUC)

	if cfg.KudaGo.Enabled {
		kudagoClient := kudagointegration.NewClient(kudagointegration.ClientConfig{
			BaseURL:        cfg.KudaGo.BaseURL,
			Location:       cfg.KudaGo.Location,
			PageSize:       cfg.KudaGo.PageSize,
			RequestTimeout: cfg.KudaGo.RequestTimeout,
		})
		kudagoSyncer := kudagointegration.NewSyncer(kudagointegration.SyncConfig{
			Enabled:      cfg.KudaGo.Enabled,
			Location:     cfg.KudaGo.Location,
			SyncInterval: cfg.KudaGo.SyncInterval,
		}, kudagoClient, pool)

		syncWG.Add(1)
		go func() {
			defer syncWG.Done()
			log.Printf("kudago sync started location=%s interval=%s", cfg.KudaGo.Location, cfg.KudaGo.SyncInterval)
			kudagoSyncer.Start(syncCtx)
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.yaml", httpx.OpenAPIYAMLHandler)
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/swagger/", httpx.SwaggerUIHandler("/openapi.yaml"))
	healthHandler := platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return nil
	})
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/auth/register", authFlowHandler.Register)
	mux.HandleFunc("/api/auth/login", authFlowHandler.Login)
	if vkOAuthCfg != nil {
		mux.HandleFunc("/api/auth/vk/login", authdelivery.VKLoginHandler(vkOAuthCfg))
		mux.HandleFunc("/api/auth/vk/callback", authdelivery.VKCallbackHandler(oauthLoginUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))
	}
	if yandexOAuthCfg != nil {
		mux.HandleFunc("/api/auth/yandex/login", authdelivery.YandexLoginHandler(yandexOAuthCfg))
		mux.HandleFunc("/api/auth/yandex/callback", authdelivery.YandexCallbackHandler(oauthLoginUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))
	}
	if googleOAuthCfg != nil {
		mux.HandleFunc("/api/auth/google/login", authdelivery.GoogleLoginHandler(googleOAuthCfg))
		mux.HandleFunc("/api/auth/google/callback", authdelivery.GoogleCallbackHandler(oauthLoginUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))
	}
	mux.HandleFunc("/api/auth/refresh", authRefreshHandler.Refresh)
	mux.HandleFunc("/api/auth/logout", authRefreshHandler.Logout)
	mux.Handle(
		"/api/me",
		platformmiddleware.AuthMiddleware(
			http.HandlerFunc(meHandler.Me),
			authdelivery.ReadAccessCookie,
			func(token string) (string, string, error) {
				claims, err := tokenService.Parse(token)
				if err != nil {
					return "", "", err
				}
				return claims.Subject, claims.Type, nil
			},
			httpx.UserIDContextKey,
		),
	)
	mux.Handle("/api/events", platformmiddleware.OptionalAuthMiddleware(
		http.HandlerFunc(placeHandler.Events),
		authdelivery.ReadAccessCookie,
		func(token string) (string, string, error) {
			claims, err := tokenService.Parse(token)
			if err != nil {
				return "", "", err
			}
			return claims.Subject, claims.Type, nil
		},
		httpx.UserIDContextKey,
	))
	mux.HandleFunc("/api/home", placeHandler.Home)
	mux.HandleFunc("/api/categories", placeHandler.Categories)
	mux.HandleFunc("/api/tags", placeHandler.Tags)
	mux.HandleFunc("/api/collections", placeHandler.Collections)
	mux.HandleFunc("/api/collections/", placeHandler.CollectionByID)
	mux.HandleFunc("/api/search", placeHandler.Search)
	mux.Handle("/api/events/", platformmiddleware.OptionalAuthMiddleware(
		http.HandlerFunc(placeHandler.EventByID),
		authdelivery.ReadAccessCookie,
		func(token string) (string, string, error) {
			claims, err := tokenService.Parse(token)
			if err != nil {
				return "", "", err
			}
			return claims.Subject, claims.Type, nil
		},
		httpx.UserIDContextKey,
	))

	cleanup := func() {
		syncCancel()
		syncWG.Wait()
		pool.Close()
	}

	return &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      platformmiddleware.RequestIDMiddleware(platformmiddleware.AccessLogMiddleware(platformmiddleware.CorsMiddleware(platformmiddleware.RecoveryMiddleware(mux)))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}, cleanup, nil
}

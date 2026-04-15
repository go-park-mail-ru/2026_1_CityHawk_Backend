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
	photonintegration "cityhawk/backend/internal/integration/photon"
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
	var placeLookupHandler *placedelivery.PlaceLookupHandler
	if cfg.Photon.Enabled {
		photonClient := photonintegration.NewClient(photonintegration.ClientConfig{
			BaseURL:        cfg.Photon.BaseURL,
			RequestTimeout: cfg.Photon.RequestTimeout,
		})
		placeLookupUC := placeusecase.NewPlaceLookupService(
			photonClient,
			placeRepo,
			[]byte(cfg.Auth.JWTSecret),
			cfg.Photon.DefaultCountry,
			cfg.Photon.DefaultTimezone,
		)
		placeLookupHandler = placedelivery.NewPlaceLookupHandler(placeLookupUC)
	}
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
	parseAccessToken := func(token string) (string, string, error) {
		claims, err := tokenService.Parse(token)
		if err != nil {
			return "", "", err
		}
		return claims.Subject, claims.Type, nil
	}
	withAuth := func(next http.HandlerFunc) http.Handler {
		return platformmiddleware.AuthMiddleware(
			http.HandlerFunc(next),
			authdelivery.ReadAccessCookie,
			parseAccessToken,
			httpx.UserIDContextKey,
		)
	}
	withOptionalAuth := func(next http.HandlerFunc) http.Handler {
		return platformmiddleware.OptionalAuthMiddleware(
			http.HandlerFunc(next),
			authdelivery.ReadAccessCookie,
			parseAccessToken,
			httpx.UserIDContextKey,
		)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /openapi.yaml", httpx.OpenAPIYAMLHandler)
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("GET /swagger/", httpx.SwaggerUIHandler("/openapi.yaml"))
	healthHandler := platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return nil
	})
	mux.HandleFunc("GET /api/health", healthHandler)
	mux.HandleFunc("POST /api/auth/register", authFlowHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authFlowHandler.Login)
	if vkOAuthCfg != nil {
		mux.HandleFunc("GET /api/auth/vk/login", authdelivery.VKLoginHandler(vkOAuthCfg))
		mux.HandleFunc("GET /api/auth/vk/callback", authdelivery.VKCallbackHandler(oauthLoginUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))
	}
	if yandexOAuthCfg != nil {
		mux.HandleFunc("GET /api/auth/yandex/login", authdelivery.YandexLoginHandler(yandexOAuthCfg))
		mux.HandleFunc("GET /api/auth/yandex/callback", authdelivery.YandexCallbackHandler(oauthLoginUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))
	}
	if googleOAuthCfg != nil {
		mux.HandleFunc("GET /api/auth/google/login", authdelivery.GoogleLoginHandler(googleOAuthCfg))
		mux.HandleFunc("GET /api/auth/google/callback", authdelivery.GoogleCallbackHandler(oauthLoginUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))
	}
	mux.HandleFunc("POST /api/auth/refresh", authRefreshHandler.Refresh)
	mux.HandleFunc("POST /api/auth/logout", authRefreshHandler.Logout)
	mux.Handle("GET /api/me", withAuth(meHandler.Me))
	mux.Handle("PATCH /api/me", withAuth(meHandler.Me))
	mux.Handle("GET /api/events", withOptionalAuth(placeHandler.Events))
	mux.Handle("POST /api/events", withAuth(placeHandler.Events))
	mux.HandleFunc("GET /api/home", placeHandler.Home)
	mux.HandleFunc("GET /api/categories", placeHandler.Categories)
	mux.HandleFunc("GET /api/tags", placeHandler.Tags)
	mux.HandleFunc("GET /api/collections", placeHandler.Collections)
	mux.HandleFunc("GET /api/collections/", placeHandler.CollectionByID)
	mux.HandleFunc("GET /api/search", placeHandler.Search)
	if placeLookupHandler != nil {
		mux.HandleFunc("GET /api/place-suggestions", placeLookupHandler.Suggestions)
		mux.Handle("POST /api/places/resolve", withAuth(placeLookupHandler.Resolve))
	}
	mux.Handle("GET /api/events/", withOptionalAuth(placeHandler.EventByID))
	mux.Handle("PATCH /api/events/", withAuth(placeHandler.EventByID))
	mux.Handle("DELETE /api/events/", withAuth(placeHandler.EventByID))

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

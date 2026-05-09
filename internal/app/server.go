package app

import (
	"context"
	"log"
	"net/http"
	"time"

	authdelivery "cityhawk/backend/internal/auth/delivery/http"
	gatewaygoogle "cityhawk/backend/internal/auth/gateway/google"
	gatewayvk "cityhawk/backend/internal/auth/gateway/vk"
	gatewayyandex "cityhawk/backend/internal/auth/gateway/yandex"
	authrepo "cityhawk/backend/internal/auth/repository"
	authusecase "cityhawk/backend/internal/auth/usecase"
	appconfig "cityhawk/backend/internal/config"
	photonintegration "cityhawk/backend/internal/integration/photon"
	placedelivery "cityhawk/backend/internal/place/delivery/http"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
	"cityhawk/backend/internal/platform/httpx"
	"cityhawk/backend/internal/platform/media"
	platformmetrics "cityhawk/backend/internal/platform/metrics"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	platformpostgres "cityhawk/backend/internal/platform/postgres"
	platformsecurity "cityhawk/backend/internal/platform/security"
	socialdelivery "cityhawk/backend/internal/social/delivery/http"
	socialrepo "cityhawk/backend/internal/social/repository"
	supportdelivery "cityhawk/backend/internal/support/delivery/http"
	supportrepo "cityhawk/backend/internal/support/repository"
	supportusecase "cityhawk/backend/internal/support/usecase"
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
	store := userrepo.NewPostgresUserRepository(pool)
	placeRepo := placerepo.NewPostgresRepository(pool)
	placeUC := placeusecase.NewService(placeRepo)
	socialRepo := socialrepo.NewPostgresRepository(pool)
	supportRepo := supportrepo.NewPostgresRepository(pool)
	supportUC := supportusecase.NewService(supportRepo, store)
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
	eventImageStore, err := media.NewLocalStorage("uploads/events", "/uploads/events")
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	placeHandler := placedelivery.NewHandler(placeUC, eventImageStore)
	authFlowHandler := authdelivery.NewAuthHandler(authFlowUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	authRefreshHandler := authdelivery.NewRefreshHandler(authUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	avatarStore, err := media.NewLocalStorage("uploads/avatars", "/uploads/avatars")
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	meHandler := userdelivery.NewMeHandler(store, avatarStore)
	supportHandler := supportdelivery.NewHandler(supportUC)
	socialHandler := socialdelivery.NewHandler(socialRepo, placeUC)
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
	withCSRF := func(next http.Handler) http.Handler {
		return platformmiddleware.CSRFMiddleware(next, authdelivery.ReadCSRFCookie)
	}
	withAuthAndCSRF := func(next http.HandlerFunc) http.Handler {
		return platformmiddleware.AuthMiddleware(
			withCSRF(http.HandlerFunc(next)),
			authdelivery.ReadAccessCookie,
			parseAccessToken,
			httpx.UserIDContextKey,
		)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", platformmetrics.Handler())
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
	mux.Handle("GET /uploads/events/", media.NewLocalFileHandler(eventImageStore.Dir(), "/uploads/events"))
	mux.Handle("GET /uploads/avatars/", media.NewLocalFileHandler(avatarStore.Dir(), "/uploads/avatars"))
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
	mux.Handle("POST /api/auth/refresh", withCSRF(http.HandlerFunc(authRefreshHandler.Refresh)))
	mux.Handle("POST /api/auth/logout", withCSRF(http.HandlerFunc(authRefreshHandler.Logout)))
	mux.Handle("GET /api/me", withAuth(meHandler.Me))
	mux.Handle("PATCH /api/me", withAuthAndCSRF(meHandler.Me))
	mux.Handle("GET /api/me/favorites", withAuth(socialHandler.Favorites))
	mux.Handle("POST /api/me/favorites/", withAuthAndCSRF(socialHandler.FavoriteByID))
	mux.Handle("DELETE /api/me/favorites/", withAuthAndCSRF(socialHandler.FavoriteByID))
	mux.Handle("GET /api/me/followers", withAuth(socialHandler.Followers))
	mux.Handle("GET /api/me/following", withAuth(socialHandler.Following))
	mux.Handle("GET /api/me/collections", withAuth(socialHandler.Collections))
	mux.Handle("POST /api/users/", withAuthAndCSRF(socialHandler.FollowByID))
	mux.Handle("DELETE /api/users/", withAuthAndCSRF(socialHandler.FollowByID))
	mux.Handle("GET /api/events", withOptionalAuth(placeHandler.Events))
	mux.Handle("POST /api/events", withAuthAndCSRF(placeHandler.Events))
	mux.HandleFunc("GET /api/home", placeHandler.Home)
	mux.HandleFunc("GET /api/categories", placeHandler.Categories)
	mux.HandleFunc("GET /api/tags", placeHandler.Tags)
	mux.HandleFunc("GET /api/cities", placeHandler.Cities)
	mux.HandleFunc("GET /api/collections", placeHandler.Collections)
	mux.HandleFunc("GET /api/collections/", placeHandler.CollectionByID)
	mux.HandleFunc("GET /api/search", placeHandler.Search)
	mux.HandleFunc("GET /api/map/collections", placeHandler.MapCollections)
	mux.HandleFunc("GET /api/map/filters", placeHandler.MapFilters)
	mux.HandleFunc("GET /api/map/collections/", placeHandler.MapCollectionSpots)
	mux.Handle("GET /api/support/tickets", withAuth(supportHandler.Tickets))
	mux.Handle("POST /api/support/tickets", withAuthAndCSRF(supportHandler.Tickets))
	mux.Handle("GET /api/support/tickets/", withAuth(supportHandler.TicketByID))
	mux.Handle("POST /api/support/tickets/", withAuthAndCSRF(supportHandler.TicketByID))
	mux.Handle("PATCH /api/support/tickets/", withAuthAndCSRF(supportHandler.TicketByID))
	mux.Handle("GET /api/support/stats", withAuth(supportHandler.Stats))
	if placeLookupHandler != nil {
		mux.HandleFunc("GET /api/place-suggestions", placeLookupHandler.Suggestions)
		mux.Handle("POST /api/places/resolve", withAuth(placeLookupHandler.Resolve))
	}
	mux.Handle("GET /api/events/", withOptionalAuth(placeHandler.EventByID))
	mux.Handle("PATCH /api/events/", withAuthAndCSRF(placeHandler.EventByID))
	mux.Handle("DELETE /api/events/", withAuthAndCSRF(placeHandler.EventByID))

	cleanup := func() {
		pool.Close()
	}

	return &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      platformmetrics.HTTPMiddleware("cityhawk-backend")(platformmiddleware.RequestIDMiddleware(platformmiddleware.AccessLogMiddleware(platformmiddleware.CorsMiddleware(platformmiddleware.CSRFTokenHeaderMiddleware(platformmiddleware.RecoveryMiddleware(mux), authdelivery.ReadCSRFCookie))))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}, cleanup, nil
}

package app

import (
	"net/http"
	"time"

	authdelivery "cityhawk/backend/internal/auth/delivery/http"
	gatewaygoogle "cityhawk/backend/internal/auth/gateway/google"
	gatewayvk "cityhawk/backend/internal/auth/gateway/vk"
	gatewayyandex "cityhawk/backend/internal/auth/gateway/yandex"
	authrepo "cityhawk/backend/internal/auth/repository"
	authusecase "cityhawk/backend/internal/auth/usecase"
	appconfig "cityhawk/backend/internal/config"
	placedelivery "cityhawk/backend/internal/place/delivery/http"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
	"cityhawk/backend/internal/platform/httpx"
	platformid "cityhawk/backend/internal/platform/id"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	platformsecurity "cityhawk/backend/internal/platform/security"
	userdelivery "cityhawk/backend/internal/user/delivery/http"
	userrepo "cityhawk/backend/internal/user/repository"
)

func NewServer(cfg appconfig.Config) *http.Server {
	store := userrepo.NewInMemoryUserRepository()
	placeRepo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	placeUC := placeusecase.NewService(placeRepo)
	refreshRepo := authrepo.NewInMemoryRefreshRepository()
	tokenService := platformsecurity.NewJWTTokenService([]byte(cfg.Auth.JWTSecret))
	authUC := authusecase.NewService(cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL, store, refreshRepo, tokenService)
	authFlowUC := authusecase.NewAuthFlowService(
		store,
		authUC,
		platformsecurity.NewBcryptPasswordService(),
		platformid.NewTimeUserIDProvider(platformid.TimeUserIDLayout),
	)
	oauthUsers := authusecase.NewOAuthUserService(
		store,
		platformsecurity.NewBcryptPasswordService(),
		platformid.NewTimeUserIDProvider(platformid.TimeUserIDLayout),
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

	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.yaml", httpx.OpenAPIYAMLHandler)
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/swagger/", httpx.SwaggerUIHandler("/openapi.yaml"))
	mux.HandleFunc("/health", platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return nil
	}))
	mux.HandleFunc("/auth/register", authFlowHandler.Register)
	mux.HandleFunc("/auth/login", authFlowHandler.Login)
	if vkOAuthCfg != nil {
		mux.HandleFunc("/auth/vk/login", authdelivery.VKLoginHandler(vkOAuthCfg))
		mux.HandleFunc("/auth/vk/callback", authdelivery.VKCallbackHandler(oauthLoginUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))
	}
	if yandexOAuthCfg != nil {
		mux.HandleFunc("/auth/yandex/login", authdelivery.YandexLoginHandler(yandexOAuthCfg))
		mux.HandleFunc("/auth/yandex/callback", authdelivery.YandexCallbackHandler(oauthLoginUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))
	}
	if googleOAuthCfg != nil {
		mux.HandleFunc("/auth/google/login", authdelivery.GoogleLoginHandler(googleOAuthCfg))
		mux.HandleFunc("/auth/google/callback", authdelivery.GoogleCallbackHandler(oauthLoginUC, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL))
	}
	mux.HandleFunc("/auth/refresh", authRefreshHandler.Refresh)
	mux.HandleFunc("/auth/logout", authRefreshHandler.Logout)
	mux.Handle(
		"/me",
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
	mux.HandleFunc("/places", placeHandler.List)
	mux.HandleFunc("/api/home", placeHandler.Home)
	mux.HandleFunc("/places/", placeHandler.Details)
	mux.HandleFunc("/places/best", placeHandler.Best)
	mux.HandleFunc("/places/category/", placeHandler.ByCategory)

	return &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      platformmiddleware.CorsMiddleware(platformmiddleware.RecoveryMiddleware(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

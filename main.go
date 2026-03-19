package main

import (
	"log"
	"net/http"
	"os"
	"time"

	authdelivery "cityhawk/backend/internal/auth/delivery/http"
	authdeliveryvalidation "cityhawk/backend/internal/auth/delivery/http/validation"
	oauthgoogle "cityhawk/backend/internal/auth/oauth/google"
	oauthvk "cityhawk/backend/internal/auth/oauth/vk"
	oauthyandex "cityhawk/backend/internal/auth/oauth/yandex"
	authrepo "cityhawk/backend/internal/auth/repository"
	authusecase "cityhawk/backend/internal/auth/usecase"
	placedelivery "cityhawk/backend/internal/place/delivery/http"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
	"cityhawk/backend/internal/platform/httpx"
	platformid "cityhawk/backend/internal/platform/id"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	platformsecurity "cityhawk/backend/internal/platform/security"
	userdelivery "cityhawk/backend/internal/user/delivery/http"
	userrepo "cityhawk/backend/internal/user/repository"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, use system env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-insecure-secret-change-me"
		log.Println("JWT_SECRET is empty, using insecure development secret")
	}

	accessTTL := parseDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute)
	refreshTTL := parseDurationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour)

	store := userrepo.NewInMemoryUserRepository()
	placeRepo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	placeUC := placeusecase.NewService(placeRepo)
	refreshRepo := authrepo.NewInMemoryRefreshRepository()
	tokenService := platformsecurity.NewJWTTokenService([]byte(secret))
	authUC := authusecase.NewService(accessTTL, refreshTTL, store, refreshRepo, tokenService)
	authFlowUC := authusecase.NewAuthFlowService(
		store,
		authUC,
		authdeliveryvalidation.NewInputValidator(),
		platformsecurity.NewBcryptPasswordService(),
		platformid.NewTimeUserIDProvider(platformid.TimeUserIDLayout),
	)
	oauthUsers := authusecase.NewOAuthUserService(
		store,
		platformsecurity.NewBcryptPasswordService(),
		platformid.NewTimeUserIDProvider(platformid.TimeUserIDLayout),
	)
	placeHandler := placedelivery.NewHandler(placeUC)
	authFlowHandler := authdelivery.NewAuthHandler(authFlowUC, accessTTL, refreshTTL)
	authRefreshHandler := authdelivery.NewRefreshHandler(authUC, accessTTL, refreshTTL)
	meHandler := userdelivery.NewMeHandler(store)
	vkOAuthCfg, err := oauthvk.NewConfigFromEnv()
	if err != nil {
		log.Printf("VK OAuth disabled: %v", err)
	}
	yandexOAuthCfg, err := oauthyandex.NewConfigFromEnv()
	if err != nil {
		log.Printf("Yandex OAuth disabled: %v", err)
	}
	googleOAuthCfg, err := oauthgoogle.NewConfigFromEnv()
	if err != nil {
		log.Printf("Google OAuth disabled: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.yaml", httpx.OpenAPIYAMLHandler)
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/swagger/", httpx.SwaggerUIHandler("/openapi.yaml"))
	mux.HandleFunc("/health", platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return nil
	}))
	mux.HandleFunc("/auth/register", authFlowHandler.Register)
	mux.HandleFunc("/auth/login", authFlowHandler.Login)
	if vkOAuthCfg != nil {
		mux.HandleFunc("/auth/vk/login", authdelivery.VKLoginHandler(vkOAuthCfg))
		mux.HandleFunc("/auth/vk/callback", authdelivery.VKCallbackHandler(vkOAuthCfg, oauthUsers, authUC, accessTTL, refreshTTL))
	}
	if yandexOAuthCfg != nil {
		mux.HandleFunc("/auth/yandex/login", authdelivery.YandexLoginHandler(yandexOAuthCfg))
		mux.HandleFunc("/auth/yandex/callback", authdelivery.YandexCallbackHandler(yandexOAuthCfg, oauthUsers, authUC, accessTTL, refreshTTL))
	}
	if googleOAuthCfg != nil {
		mux.HandleFunc("/auth/google/login", authdelivery.GoogleLoginHandler(googleOAuthCfg))
		mux.HandleFunc("/auth/google/callback", authdelivery.GoogleCallbackHandler(googleOAuthCfg, oauthUsers, authUC, accessTTL, refreshTTL))
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

	server := http.Server{
		Addr:         ":" + port,
		Handler:      platformmiddleware.CorsMiddleware(platformmiddleware.RecoveryMiddleware(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("server started on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

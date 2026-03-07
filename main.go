package main

import (
	"log"
	"net/http"
	"os"
	"time"

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

	store := &userStore{byEmail: make(map[string]user)}
	auth := &authService{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		refreshSessions: &refreshStore{
			session: make(map[string]refreshSession),
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /auth/register", registerHandler(store))
	mux.HandleFunc("POST /auth/login", loginHandler(store, auth))
	mux.HandleFunc("POST /auth/refresh", refreshHandler(auth))
	mux.HandleFunc("POST /auth/logout", logoutHandler(auth))
	mux.Handle("GET /me", authMiddleware(auth, meHandler(store)))

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("server started on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

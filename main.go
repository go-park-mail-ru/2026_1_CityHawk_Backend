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
	places := newPlaceStore()
	auth := &authService{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		refreshSessions: &refreshStore{
			session: make(map[string]refreshSession),
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/auth/register", registerHandler(store, auth))
	mux.HandleFunc("/auth/login", loginHandler(store, auth))
	mux.HandleFunc("/auth/refresh", refreshHandler(store, auth))
	mux.HandleFunc("/auth/logout", logoutHandler(auth))
	mux.Handle("/me", authMiddleware(auth, meHandler(store)))
	mux.HandleFunc("/api/home", homeHandler())
	mux.HandleFunc("/places", placesListHandler(places))
	mux.HandleFunc("/api/home", homeHandler(places))
	mux.HandleFunc("/places/", placeDetailsHandler(places))
	mux.HandleFunc("/places/best", placesBestHandler(places))
	mux.HandleFunc("/places/category/", placesByCategoryListHandler(places))

	server := http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("server started on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

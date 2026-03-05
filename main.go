package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type user struct {
	ID           string
	Email        string
	PasswordHash string
}

type userStore struct {
	mu      sync.RWMutex
	byEmail map[string]user
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, use system env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := &userStore{byEmail: make(map[string]user)}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /auth/register", registerHandler(store))
	mux.HandleFunc("POST /auth/login", loginHandler(store))

	server := http.Server{
		Addr: ":"+port,
		Handler: mux,
	}

	log.Printf("server started on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func registerHandler(store *userStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}

		email := strings.TrimSpace(strings.ToLower(req.Email))
		password := strings.TrimSpace(req.Password)
		if email == "" || password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
			return
		}

		u, err := createUser(email, password)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		if err := store.create(u); err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]string{
			"id":    u.ID,
			"email": u.Email,
		})
	}
}

func loginHandler(store *userStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}

		email := strings.TrimSpace(strings.ToLower(req.Email))
		password := strings.TrimSpace(req.Password)
		if email == "" || password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
			return
		}

		u, ok := store.getByEmail(email)
		if !ok || !verifyPassword(password, u.PasswordHash) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "login successful"})
	}
}

func (s *userStore) create(u user) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byEmail[u.Email]; exists {
		return errors.New("email already exists")
	}
	s.byEmail[u.Email] = u
	return nil
}

func (s *userStore) getByEmail(email string) (user, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byEmail[email]
	return u, ok
}

func createUser(email, password string) (user, error) {
	salt, err := randomHex(16)
	if err != nil {
		return user{}, err
	}
	return user{
		ID:           time.Now().UTC().Format("20060102150405.000000000"),
		Email:        email,
		PasswordHash: hashPassword(password, salt),
	}, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return salt + ":" + hex.EncodeToString(sum[:])
}

func verifyPassword(password, stored string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return false
	}
	return hashPassword(password, parts[0]) == stored
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

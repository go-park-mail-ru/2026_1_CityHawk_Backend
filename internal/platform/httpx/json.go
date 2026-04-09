package httpx

import (
	"encoding/json"
	"net/http"
)

type ContextKey string

const UserIDContextKey ContextKey = "userID"
const RequestIDContextKey ContextKey = "requestID"
const RequestIDHeader = "X-Request-ID"

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

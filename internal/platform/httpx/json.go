package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/mailru/easyjson"
)

type ContextKey string

const UserIDContextKey ContextKey = "userID"
const RequestIDContextKey ContextKey = "requestID"
const RequestIDHeader = "X-Request-ID"

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if marshaler, ok := payload.(easyjson.Marshaler); ok {
		if data, err := easyjson.Marshal(marshaler); err == nil {
			_, _ = w.Write(append(data, '\n'))
			return
		}
	}
	_ = json.NewEncoder(w).Encode(payload)
}

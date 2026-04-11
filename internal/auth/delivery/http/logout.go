package http

import (
	"net/http"

	"cityhawk/backend/internal/platform/httpx"
)

func (h *RefreshHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken := ReadRefreshCookie(r)
	if refreshToken == "" {
		httpx.WriteJSON(w, http.StatusUnauthorized, httpx.NewErrorResponse("Session expired", nil))
		return
	}

	if err := h.refreshUC.RevokeRefresh(r.Context(), refreshToken); err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, httpx.NewErrorResponse("Session expired", nil))
		return
	}
	ClearRefreshCookie(w)
	ClearAccessCookie(w)
	httpx.WriteJSON(w, http.StatusOK, okResponse{OK: true})
}

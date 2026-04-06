package http

import (
	"net/http"

	"cityhawk/backend/internal/platform/httpx"
)

func (h *RefreshHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := ReadRefreshCookie(r)
	if refreshToken == "" {
		httpx.WriteJSON(w, http.StatusUnauthorized, httpx.NewErrorResponse("missing refresh token", nil))
		return
	}

	resp, err := h.refreshUC.RotateRefresh(r.Context(), refreshToken)
	if err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, httpx.NewErrorResponse("invalid refresh token", nil))
		return
	}

	SetRefreshCookie(w, resp.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.AccessToken, h.accessTTL)
	httpx.WriteJSON(w, http.StatusOK, refreshResponse{AccessToken: resp.AccessToken})
}

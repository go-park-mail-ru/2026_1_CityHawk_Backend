package http

import (
	"net/http"

	"cityhawk/backend/internal/platform/httpx"
)

func (h *RefreshHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken := ReadRefreshCookie(r)
	if refreshToken == "" {
		httpx.WriteJSON(w, http.StatusUnauthorized, httpx.NewErrorResponse("missing refresh token", nil))
		return
	}

	if err := h.refreshUC.RevokeRefresh(r.Context(), refreshToken); err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, httpx.NewErrorResponse("internal error", nil))
		return
	}
	ClearRefreshCookie(w)
	ClearAccessCookie(w)
	httpx.WriteJSON(w, http.StatusOK, messageResponse{Message: "logout successful"})
}

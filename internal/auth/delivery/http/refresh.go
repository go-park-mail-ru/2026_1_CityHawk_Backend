package http

import (
	"net/http"

	"cityhawk/backend/internal/platform/httpx"
)

func (h *RefreshHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := ReadRefreshCookie(r)
	if refreshToken == "" {
		httpx.WriteJSON(w, http.StatusUnauthorized, httpx.NewErrorResponse("Session expired", nil))
		return
	}

	resp, err := h.refreshUC.RotateRefresh(r.Context(), refreshToken)
	if err != nil {
		httpx.WriteJSON(w, http.StatusUnauthorized, httpx.NewErrorResponse("Session expired", nil))
		return
	}

	SetRefreshCookie(w, resp.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.AccessToken, h.accessTTL)
	csrfToken, err := IssueCSRFToken()
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, httpx.NewErrorResponse("failed to issue csrf token", nil))
		return
	}
	SetCSRFCookie(w, csrfToken, h.refreshTTL)
	httpx.WriteJSON(w, http.StatusOK, okResponse{OK: true})
}

package http

import "net/http"

func (h *RefreshHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := ReadRefreshCookie(r)
	if refreshToken == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
		return
	}

	resp, err := h.refreshUC.RotateRefresh(refreshToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		return
	}

	SetRefreshCookie(w, resp.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.AccessToken, h.accessTTL)
	writeJSON(w, http.StatusOK, map[string]string{"access_token": resp.AccessToken})
}

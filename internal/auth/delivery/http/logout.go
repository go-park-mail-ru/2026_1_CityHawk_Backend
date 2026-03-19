package http

import "net/http"

func (h *RefreshHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken := ReadRefreshCookie(r)
	if refreshToken == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
		return
	}

	h.refreshUC.RevokeRefresh(refreshToken)
	ClearRefreshCookie(w)
	ClearAccessCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logout successful"})
}

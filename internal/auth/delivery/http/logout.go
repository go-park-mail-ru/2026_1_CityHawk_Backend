package http

import "net/http"

func (h *RefreshHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken := ReadRefreshCookie(r)
	if refreshToken == "" {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "missing refresh token"})
		return
	}

	if err := h.refreshUC.RevokeRefresh(r.Context(), refreshToken); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}
	ClearRefreshCookie(w)
	ClearAccessCookie(w)
	writeJSON(w, http.StatusOK, messageResponse{Message: "logout successful"})
}

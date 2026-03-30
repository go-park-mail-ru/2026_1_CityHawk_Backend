package httpx

import (
	"errors"
	"net/http"
)

type HTTPError struct {
	status  int
	message string
}

func (e HTTPError) Error() string {
	return e.message
}

func NewHTTPError(status int, message string) error {
	return HTTPError{status: status, message: message}
}

func WriteMappedError(w http.ResponseWriter, err error) {
	var he HTTPError
	if errors.As(err, &he) {
		WriteJSON(w, he.status, map[string]string{"error": he.message})
		return
	}
	WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

package httpx

import (
	"errors"
	"net/http"
)

type HTTPError struct {
	status  int
	message string
	details any
}

func (e HTTPError) Error() string {
	return e.message
}

type ErrorResponse struct {
	Error   string         `json:"error"`
	Details map[string]any `json:"details"`
}

func NewHTTPError(status int, message string) error {
	return HTTPError{status: status, message: message}
}

func NewHTTPErrorWithDetails(status int, message string, details any) error {
	return HTTPError{status: status, message: message, details: details}
}

func NewErrorResponse(message string, details any) ErrorResponse {
	return ErrorResponse{
		Error:   message,
		Details: normalizeErrorDetails(details),
	}
}

func WriteMappedError(w http.ResponseWriter, err error) {
	var he HTTPError
	if errors.As(err, &he) {
		WriteJSON(w, he.status, NewErrorResponse(he.message, he.details))
		return
	}
	WriteJSON(w, http.StatusInternalServerError, NewErrorResponse("internal error", nil))
}

func normalizeErrorDetails(details any) map[string]any {
	switch v := details.(type) {
	case nil:
		return map[string]any{}
	case map[string]any:
		if v == nil {
			return map[string]any{}
		}
		return v
	case map[string]string:
		out := make(map[string]any, len(v))
		for k, val := range v {
			out[k] = val
		}
		return out
	default:
		return map[string]any{"value": v}
	}
}

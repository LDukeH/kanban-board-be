package apperrors

import (
	"errors"
	"net/http"
)

// WriteError maps domain errors to HTTP status codes and writes the response.
func WriteError(w http.ResponseWriter, err error) {
	var status int

	switch {
	case errors.Is(err, ErrValidation):
		status = http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrInvalidCredential):
		status = http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrConflict):
		status = http.StatusConflict
	default:
		status = http.StatusInternalServerError
	}

	http.Error(w, err.Error(), status)
}
package apperrors

import "errors"

// Authentication
var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInvalidCredential = errors.New("invalid credentials")
)

// Authorization
var (
	ErrForbidden = errors.New("forbidden")
)

// Resources
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

// Validation
var (
	ErrValidation = errors.New("validation failed")
)

// Internal
var (
	ErrInternal = errors.New("internal error")
)
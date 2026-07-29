package authz

import "errors"

var (
	// ErrForbidden is returned when the user does not have the required role.
	ErrForbidden = errors.New("forbidden")

	// ErrNotMember is returned when the user is not a member of the workspace.
	ErrNotMember = errors.New("user is not a member of this workspace")
)
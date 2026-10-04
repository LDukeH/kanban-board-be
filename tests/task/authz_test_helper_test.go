package task_test

import (
	"testing"

	"kanban-board-be/internal/feature/authz"
)

// newAuthzServiceForTest builds the same *authz.Service the production server
// wires, but backed by a stub Resolver so service tests need no database.
func newAuthzServiceForTest(t *testing.T, resolver authz.Resolver) *authz.Service {
	t.Helper()
	return authz.NewService(resolver)
}

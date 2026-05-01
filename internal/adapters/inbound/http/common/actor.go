package common

import "github.com/google/uuid"

// Actor represents the authenticated user making a request
// This is extracted from X-User-ID header and enriched with user details
type Actor struct {
	UserID     uuid.UUID // Internal UUID used throughout the system
	ExternalID string    // External ID from X-User-ID header (for logging/audit)
	Role       string    // User role (for authorization checks)
}

// contextKey is a private type for context keys to avoid collisions
type contextKey string

// actorContextKey is the key used to store Actor in request context
const actorContextKey contextKey = "actor"

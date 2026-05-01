package shared

import (
	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/common"
	"github.com/labstack/echo/v4"
)

// Actor represents request-scoped user info for application layer
// This is the application-layer version (string UUID)
// Converted from HTTP-layer Actor (uuid.UUID)
type Actor struct {
	UserID string
	Role   string // User role for authorization
}

// ActorFromEcho extracts Actor from echo context and converts to application-layer Actor
// The middleware has already validated the user and enriched the context
// This function reads from context and converts UUID to string for application layer
func ActorFromEcho(c echo.Context) (Actor, error) {
	// Get HTTP-layer actor from context (has uuid.UUID)
	httpActor, err := common.GetActor(c)
	if err != nil {
		return Actor{}, err
	}

	// Convert to application-layer actor (string UUID)
	return Actor{
		UserID: httpActor.UserID.String(),
	}, nil
}

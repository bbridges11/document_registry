package middleware

import (
	"net/http"

	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/common"
	"github.com/bbridges_11/document-registry/internal/application/user"
	domainUser "github.com/bbridges_11/document-registry/internal/domain/user"
	"github.com/labstack/echo/v4"
)

// UserValidationMiddleware validates that a user exists and enriches context with Actor
type UserValidationMiddleware struct {
	userQueries user.QueryService
}

func NewUserValidationMiddleware(userQueries user.QueryService) *UserValidationMiddleware {
	return &UserValidationMiddleware{
		userQueries: userQueries,
	}
}

// ValidateUserExists middleware:
// 1. Validates X-User-ID header format (7-char external ID)
// 2. Looks up user by external ID
// 3. Enriches request context with Actor (internal UUID + role)
// 4. Handlers can then extract Actor from context
func (m *UserValidationMiddleware) ValidateUserExists() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			externalID := c.Request().Header.Get("X-User-ID")
			if externalID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "X-User-ID header is required",
				})
			}

			// Validate external ID format (7 alphanumeric chars)
			if err := domainUser.ValidateExternalID(externalID); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid external ID format: must be 7 alphanumeric characters",
				})
			}

			// Normalize to lowercase
			normalizedID := domainUser.NormalizeExternalID(externalID)

			// Lookup user by external ID
			userDTO, err := m.userQueries.GetUserByExternalID(c.Request().Context(), user.GetUserByExternalIDQuery{
				ExternalID: normalizedID,
			})
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "user not found or inactive",
				})
			}

			// Create Actor with internal UUID
			actor := common.Actor{
				UserID:     userDTO.ID,         // Internal UUID
				ExternalID: userDTO.ExternalID, // External ID (for logging)
				Role:       userDTO.Role,       // For authorization
			}

			// Enrich context with Actor
			ctx := common.WithActor(c.Request().Context(), actor)
			c.SetRequest(c.Request().WithContext(ctx))

			// Continue to handler with enriched context
			return next(c)
		}
	}
}

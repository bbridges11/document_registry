package middleware

import (
	"net/http"

	"github.com/bbridges_11/document-registry/internal/application/user"
	domainUser "github.com/bbridges_11/document-registry/internal/domain/user"
	"github.com/labstack/echo/v4"
)

// UserValidationMiddleware validates that a user exists before allowing operations
type UserValidationMiddleware struct {
	userQueries user.QueryService
}

func NewUserValidationMiddleware(userQueries user.QueryService) *UserValidationMiddleware {
	return &UserValidationMiddleware{
		userQueries: userQueries,
	}
}

// ValidateUserExists middleware checks if the X-User-ID header contains a valid, active user
func (m *UserValidationMiddleware) ValidateUserExists() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			externalID := c.Request().Header.Get("X-User-ID")
			if externalID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "X-User-ID header is required",
				})
			}

			// Validate external ID format
			if err := domainUser.ValidateExternalID(externalID); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid external ID format",
				})
			}

			// Normalize external ID
			normalizedID := domainUser.NormalizeExternalID(externalID)

			// Check if user exists by external ID
			_, err := m.userQueries.GetUserByExternalID(c.Request().Context(), user.GetUserByExternalIDQuery{
				ExternalID: normalizedID,
			})
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "user not found or inactive",
				})
			}

			// User exists and is active, continue
			return next(c)
		}
	}
}

package middleware

import (
	"net/http"

	"github.com/bbridges_11/document-registry/internal/application/user"
	domainUser "github.com/bbridges_11/document-registry/internal/domain/user"
	"github.com/labstack/echo/v4"
)

// AdminCheckMiddleware validates that the authenticated user has admin role
type AdminCheckMiddleware struct {
	userQueries user.QueryService
}

func NewAdminCheckMiddleware(userQueries user.QueryService) *AdminCheckMiddleware {
	return &AdminCheckMiddleware{
		userQueries: userQueries,
	}
}

// RequireAdmin middleware checks if the user has admin role
// NOTE: This middleware should be used AFTER UserValidation middleware
// as it assumes X-User-ID has already been validated
func (m *AdminCheckMiddleware) RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			externalID := c.Request().Header.Get("X-User-ID")
			if externalID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "X-User-ID header is required",
				})
			}

			// Validate and normalize external ID
			if err := domainUser.ValidateExternalID(externalID); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid external ID format",
				})
			}

			normalizedID := domainUser.NormalizeExternalID(externalID)

			// Get user details by external ID
			userDTO, err := m.userQueries.GetUserByExternalID(c.Request().Context(), user.GetUserByExternalIDQuery{
				ExternalID: normalizedID,
			})
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to retrieve user details",
				})
			}

			// Check if user has admin role
			if userDTO.Role != "admin" {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "admin privileges required",
				})
			}

			// User is admin, continue
			return next(c)
		}
	}
}

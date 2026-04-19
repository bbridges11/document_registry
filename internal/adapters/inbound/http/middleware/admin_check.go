package middleware

import (
	"net/http"

	"github.com/bbridges_11/document-registry/internal/application/user"
	"github.com/google/uuid"
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
			userIDStr := c.Request().Header.Get("X-User-ID")
			if userIDStr == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "X-User-ID header is required",
				})
			}

			// Parse user ID
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid user ID format",
				})
			}

			// Get user details
			userDTO, err := m.userQueries.GetUser(c.Request().Context(), user.GetUserQuery{
				ID: userID,
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

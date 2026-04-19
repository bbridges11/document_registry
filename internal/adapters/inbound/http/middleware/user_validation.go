package middleware

import (
	"net/http"

	"github.com/bbridges_11/document-registry/internal/application/user"
	"github.com/google/uuid"
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
			userIDStr := c.Request().Header.Get("X-User-ID")
			if userIDStr == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "X-User-ID header is required",
				})
			}

			// Parse user ID as UUID
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid user ID format",
				})
			}

			// Check if user exists and is active
			exists, err := m.userQueries.UserExists(c.Request().Context(), userID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to validate user",
				})
			}

			if !exists {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "user not found or inactive",
				})
			}

			// User exists and is active, continue
			return next(c)
		}
	}
}

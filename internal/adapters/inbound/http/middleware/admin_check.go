package middleware

import (
	"net/http"

	"github.com/bbridges_11/document-registry/internal/adapters/inbound/http/common"
	"github.com/labstack/echo/v4"
)

// AdminCheckMiddleware validates that the authenticated user has admin role
type AdminCheckMiddleware struct{}

func NewAdminCheckMiddleware() *AdminCheckMiddleware {
	return &AdminCheckMiddleware{}
}

// RequireAdmin middleware checks if the user has admin role
// NOTE: This middleware MUST be used AFTER UserValidationMiddleware
// as it reads Actor from the request context
func (m *AdminCheckMiddleware) RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract Actor from context (already validated and enriched by UserValidationMiddleware)
			actor, err := common.GetActor(c)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "internal error: actor not found in context",
				})
			}

			// Check if user has admin role
			if actor.Role != "admin" {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "admin privileges required",
				})
			}

			// User is admin, continue
			return next(c)
		}
	}
}

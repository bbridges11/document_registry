package validation

import (
	"encoding/base64"
	"net/http"

	httpshared "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/shared"
	appValidation "github.com/bbridges_11/document-registry/internal/application/validation"
	"github.com/labstack/echo/v4"
)

// Handler handles validation HTTP requests
type Handler struct {
	validationService appValidation.UseCase
}

// NewHandler creates a new validation handler
func NewHandler(validationService appValidation.UseCase) *Handler {
	return &Handler{
		validationService: validationService,
	}
}

// MiddlewareConfig contains middleware for validation routes
type MiddlewareConfig struct {
	UserValidation echo.MiddlewareFunc
}

// RegisterRoutes registers validation routes
func (h *Handler) RegisterRoutes(e *echo.Echo, mw *MiddlewareConfig) {
	middlewares := []echo.MiddlewareFunc{}
	if mw != nil && mw.UserValidation != nil {
		middlewares = append(middlewares, mw.UserValidation)
	}

	e.POST("/validation/content", h.ValidateContent, middlewares...)
}

// ValidateContent handles document content validation
func (h *Handler) ValidateContent(c echo.Context) error {
	// Extract actor
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Bind JSON request
	var req ValidateContentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	// Validate required fields
	if req.DocumentType == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "document_type is required"})
	}

	// Decode base64 content
	contentBytes, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "content must be valid base64 encoded"})
	}

	// Map to application input
	input := toValidateContentInput(
		appValidation.Actor{UserID: actor.UserID},
		contentBytes,
		req.DocumentType,
	)

	// Call application service
	output, err := h.validationService.ValidateContent(c.Request().Context(), input)
	if err != nil {
		return handleError(c, err)
	}

	// Map to response
	response := toValidateContentResponse(output)

	// Return appropriate status code
	if !output.Valid {
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	return c.JSON(http.StatusOK, response)
}

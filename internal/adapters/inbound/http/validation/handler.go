package validation

import (
	"fmt"
	"io"
	"net/http"

	httpshared "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/shared"
	appValidation "github.com/bbridges_11/document-registry/internal/application/validation"
	"github.com/labstack/echo/v4"
)

const maxContentSize = 10 * 1024 * 1024 // 10MB

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

	e.POST("/validation/definition", h.ValidateDefinition, middlewares...)
}

// ValidateDefinition handles definition file validation
func (h *Handler) ValidateDefinition(c echo.Context) error {
	// Extract actor
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid multipart form"})
	}

	// Get file from form
	files := form.File["file"]
	if len(files) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file field required"})
	}

	// Open file
	file, err := files[0].Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to open file"})
	}
	defer file.Close()

	// Read content with size limit
	limitReader := io.LimitReader(file, maxContentSize+1)
	contentBytes, err := io.ReadAll(limitReader)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to read file"})
	}

	// Check size limit
	if len(contentBytes) > maxContentSize {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("file exceeds maximum size of %d bytes", maxContentSize),
		})
	}

	// Get document type from form
	documentType := "definition" // Default
	if docTypes := form.Value["document_type"]; len(docTypes) > 0 {
		documentType = docTypes[0]
	}

	// Map to application input
	input := toValidateDefinitionInput(
		appValidation.Actor{UserID: actor.UserID},
		contentBytes,
		documentType,
	)

	// Call application service
	output, err := h.validationService.ValidateDefinition(c.Request().Context(), input)
	if err != nil {
		return handleError(c, err)
	}

	// Map to response
	response := toValidateDefinitionResponse(output)

	// Return appropriate status code
	if !output.Valid {
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	return c.JSON(http.StatusOK, response)
}

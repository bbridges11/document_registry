package deprecation

import (
	"net/http"

	httpshared "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/shared"
	appdep "github.com/bbridges_11/document-registry/internal/application/deprecation"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for deprecation operations
type Handler struct {
	service appdep.UseCase
}

// NewHandler creates a new deprecation handler
func NewHandler(service appdep.UseCase) *Handler {
	return &Handler{service: service}
}

// MiddlewareConfig holds middleware functions
type MiddlewareConfig struct {
	UserValidation echo.MiddlewareFunc
	Authorization  echo.MiddlewareFunc
}

// RegisterRoutes registers deprecation routes
func (h *Handler) RegisterRoutes(e *echo.Echo, mw *MiddlewareConfig) {
	protectedMW := []echo.MiddlewareFunc{}
	if mw != nil {
		if mw.UserValidation != nil {
			protectedMW = append(protectedMW, mw.UserValidation)
		}
		if mw.Authorization != nil {
			protectedMW = append(protectedMW, mw.Authorization)
		}
	}

	// All deprecation endpoints require authentication
	e.POST("/versions/:id/deprecate", h.RequestDeprecation, protectedMW...)
	e.POST("/versions/:id/deprecation/approve", h.ApproveDeprecation, protectedMW...)
	e.POST("/versions/:id/deprecation/reject", h.RejectDeprecation, protectedMW...)
	e.POST("/versions/:id/deprecation/cancel", h.CancelDeprecation, protectedMW...)
	e.GET("/versions/:id/deprecation", h.GetDeprecation, protectedMW...)
	e.GET("/versions/:id/deprecation/status", h.GetDeprecationStatus, protectedMW...)
}

// RequestDeprecation handles POST /versions/:id/deprecate
func (h *Handler) RequestDeprecation(c echo.Context) error {
	// Parse version ID from path
	versionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}

	// Parse request body
	var req RequestDeprecationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Validate required fields
	if req.Reason == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "reason is required"})
	}

	// Get actor from context
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Call service
	output, err := h.service.RequestDeprecation(c.Request().Context(), appdep.RequestDeprecationInput{
		Actor:           appdep.Actor{UserID: actor.UserID},
		VersionID:       versionID,
		Reason:          req.Reason,
		DeprecationNote: req.DeprecationNote,
	})
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusCreated, toDeprecationResponse(output))
}

// ApproveDeprecation handles POST /versions/:id/deprecation/approve
func (h *Handler) ApproveDeprecation(c echo.Context) error {
	// Parse version ID from path
	versionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}

	// Parse request body
	var req ApproveDeprecationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Validate required fields
	if req.Role == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "role is required"})
	}

	// Get actor from context
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Call service
	err = h.service.ApproveDeprecation(c.Request().Context(), appdep.ApproveDeprecationInput{
		Actor:     appdep.Actor{UserID: actor.UserID},
		VersionID: versionID,
		Role:      req.Role,
		Comment:   req.Comment,
	})
	if err != nil {
		return handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RejectDeprecation handles POST /versions/:id/deprecation/reject
func (h *Handler) RejectDeprecation(c echo.Context) error {
	// Parse version ID from path
	versionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}

	// Parse request body
	var req RejectDeprecationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Validate required fields
	if req.Reason == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "reason is required"})
	}

	// Get actor from context
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Call service
	err = h.service.RejectDeprecation(c.Request().Context(), appdep.RejectDeprecationInput{
		Actor:     appdep.Actor{UserID: actor.UserID},
		VersionID: versionID,
		Reason:    req.Reason,
	})
	if err != nil {
		return handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// CancelDeprecation handles POST /versions/:id/deprecation/cancel
func (h *Handler) CancelDeprecation(c echo.Context) error {
	// Parse version ID from path
	versionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}

	// Get actor from context
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Call service
	err = h.service.CancelDeprecation(c.Request().Context(), appdep.CancelDeprecationInput{
		Actor:     appdep.Actor{UserID: actor.UserID},
		VersionID: versionID,
	})
	if err != nil {
		return handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// GetDeprecation handles GET /versions/:id/deprecation
func (h *Handler) GetDeprecation(c echo.Context) error {
	// Parse version ID from path
	versionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}

	// Get actor from context
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Call service
	output, err := h.service.GetDeprecation(c.Request().Context(), appdep.GetDeprecationInput{
		Actor:     appdep.Actor{UserID: actor.UserID},
		VersionID: versionID,
	})
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, toDeprecationResponse(output))
}

// GetDeprecationStatus handles GET /versions/:id/deprecation/status
func (h *Handler) GetDeprecationStatus(c echo.Context) error {
	// Parse version ID from path
	versionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}

	// Get actor from context
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Call service
	output, err := h.service.GetDeprecationStatus(c.Request().Context(), appdep.GetDeprecationInput{
		Actor:     appdep.Actor{UserID: actor.UserID},
		VersionID: versionID,
	})
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, toDeprecationStatusResponse(output))
}

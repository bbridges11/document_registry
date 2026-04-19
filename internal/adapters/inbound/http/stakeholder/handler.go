package stakeholder

import (
	"net/http"

	httpshared "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/shared"
	appstakeholder "github.com/bbridges_11/document-registry/internal/application/stakeholder"
	domainStakeholder "github.com/bbridges_11/document-registry/internal/domain/stakeholder"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct{ stakeholders appstakeholder.Service }

func NewHandler(stakeholders appstakeholder.Service) *Handler {
	return &Handler{stakeholders: stakeholders}
}

type MiddlewareConfig struct{ UserValidation echo.MiddlewareFunc }

func (h *Handler) RegisterRoutes(e *echo.Echo, mw *MiddlewareConfig) {
	middlewares := []echo.MiddlewareFunc{}
	if mw != nil && mw.UserValidation != nil {
		middlewares = append(middlewares, mw.UserValidation)
	}
	e.POST("/documents/:documentId/stakeholders", h.AddStakeholder, middlewares...)
	e.GET("/documents/:documentId/stakeholders", h.ListStakeholders, middlewares...)
	e.DELETE("/documents/:documentId/stakeholders/:userId", h.RemoveStakeholder, middlewares...)
}
func (h *Handler) AddStakeholder(c echo.Context) error {
	documentID, err := uuid.Parse(c.Param("documentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid document ID"})
	}
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	var req AddStakeholderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}
	if err := h.stakeholders.Add(c.Request().Context(), appstakeholder.AddInput{Actor: appstakeholder.Actor{UserID: actor.UserID}, DocumentID: documentID, UserID: req.UserID, Role: domainStakeholder.Role(req.Role)}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusCreated)
}
func (h *Handler) ListStakeholders(c echo.Context) error {
	documentID, err := uuid.Parse(c.Param("documentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid document ID"})
	}
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	views, err := h.stakeholders.List(c.Request().Context(), appstakeholder.ListInput{Actor: appstakeholder.Actor{UserID: actor.UserID}, DocumentID: documentID})
	if err != nil {
		return handleError(c, err)
	}
	responses := make([]StakeholderResponse, 0, len(views))
	for _, v := range views {
		responses = append(responses, StakeholderResponse{DocumentID: v.DocumentID, UserID: v.UserID, UserName: v.UserName, Role: v.Role, CreatedAt: v.CreatedAt})
	}
	return c.JSON(http.StatusOK, responses)
}
func (h *Handler) RemoveStakeholder(c echo.Context) error {
	documentID, err := uuid.Parse(c.Param("documentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid document ID"})
	}
	stakeholderUserID := c.Param("userId")
	if stakeholderUserID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user ID required"})
	}
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	if err := h.stakeholders.Remove(c.Request().Context(), appstakeholder.RemoveInput{Actor: appstakeholder.Actor{UserID: actor.UserID}, DocumentID: documentID, UserID: stakeholderUserID}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

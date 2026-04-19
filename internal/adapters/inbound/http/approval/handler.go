package approval

import (
	"net/http"

	httpshared "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/shared"
	appapproval "github.com/bbridges_11/document-registry/internal/application/approval"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct{ approvals appapproval.Service }

func NewHandler(approvals appapproval.Service) *Handler { return &Handler{approvals: approvals} }

type MiddlewareConfig struct{ UserValidation echo.MiddlewareFunc }

func (h *Handler) RegisterRoutes(e *echo.Echo, mw *MiddlewareConfig) {
	middlewares := []echo.MiddlewareFunc{}
	if mw != nil && mw.UserValidation != nil {
		middlewares = append(middlewares, mw.UserValidation)
	}
	e.POST("/versions/:versionId/approvals/grant", h.GrantApproval, middlewares...)
	e.POST("/versions/:versionId/approvals/revoke/:userId", h.RevokeApproval, middlewares...)
	e.GET("/approvals/:id", h.GetApproval, middlewares...)
	e.GET("/versions/:versionId/approvals/list", h.ListApprovalsByVersion, middlewares...)
	e.GET("/versions/:versionId/approvals/summary", h.GetApprovalSummary, middlewares...)
}
func (h *Handler) GrantApproval(c echo.Context) error {
	versionID, err := uuid.Parse(c.Param("versionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	var req GrantApprovalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}
	if err := h.approvals.Grant(c.Request().Context(), appapproval.GrantInput{VersionID: versionID, UserID: actor.UserID, Role: req.Role, Comment: req.Comment}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusCreated)
}
func (h *Handler) RevokeApproval(c echo.Context) error {
	versionID, err := uuid.Parse(c.Param("versionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user ID required"})
	}
	if _, err := httpshared.ActorFromEcho(c); err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	if err := h.approvals.Revoke(c.Request().Context(), appapproval.RevokeInput{VersionID: versionID, UserID: targetUserID}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *Handler) GetApproval(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid approval ID"})
	}
	view, err := h.approvals.Get(c.Request().Context(), appapproval.GetInput{ID: id})
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, ApprovalResponse{ID: view.ID, VersionID: view.VersionID, UserID: view.UserID, UserName: view.UserName, Role: view.Role, Approved: view.Approved, Comment: view.Comment, CreatedAt: view.CreatedAt})
}
func (h *Handler) ListApprovalsByVersion(c echo.Context) error {
	versionID, err := uuid.Parse(c.Param("versionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	views, err := h.approvals.ListByVersion(c.Request().Context(), appapproval.ListByVersionInput{VersionID: versionID})
	if err != nil {
		return handleError(c, err)
	}
	responses := make([]ApprovalResponse, 0, len(views))
	for _, v := range views {
		responses = append(responses, ApprovalResponse{ID: v.ID, VersionID: v.VersionID, UserID: v.UserID, UserName: v.UserName, Role: v.Role, Approved: v.Approved, Comment: v.Comment, CreatedAt: v.CreatedAt})
	}
	return c.JSON(http.StatusOK, responses)
}
func (h *Handler) GetApprovalSummary(c echo.Context) error {
	versionID, err := uuid.Parse(c.Param("versionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	view, err := h.approvals.GetSummary(c.Request().Context(), versionID)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, ApprovalSummaryResponse{VersionID: view.VersionID, Required: view.Required, Received: view.Received, Remaining: view.Remaining, Complete: view.Complete})
}

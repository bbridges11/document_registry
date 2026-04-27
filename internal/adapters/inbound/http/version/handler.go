package version

import (
	"bytes"
	"encoding/base64"
	"net/http"

	httpshared "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/shared"
	appver "github.com/bbridges_11/document-registry/internal/application/version"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct{ versions appver.UseCase }

func NewHandler(versions appver.UseCase) *Handler { return &Handler{versions: versions} }

type MiddlewareConfig struct {
	UserValidation echo.MiddlewareFunc
	Authorization  echo.MiddlewareFunc
}

func (h *Handler) RegisterRoutes(e *echo.Echo, mw *MiddlewareConfig) {
	standardMW := []echo.MiddlewareFunc{}
	if mw != nil && mw.UserValidation != nil {
		standardMW = append(standardMW, mw.UserValidation)
	}
	protectedMW := []echo.MiddlewareFunc{}
	if mw != nil {
		if mw.UserValidation != nil {
			protectedMW = append(protectedMW, mw.UserValidation)
		}
		if mw.Authorization != nil {
			protectedMW = append(protectedMW, mw.Authorization)
		}
	}

	e.POST("/documents/:documentId/versions", h.CreateVersion, standardMW...)
	e.GET("/documents/:documentId/versions", h.ListVersionsByDocument, standardMW...)
	e.GET("/versions/:id", h.GetVersion, standardMW...)
	e.PUT("/versions/:id", h.UpdateVersion, protectedMW...)
	e.POST("/versions/:id/submit", h.SubmitVersion, protectedMW...)
	e.POST("/versions/:id/review", h.ReviewVersion, protectedMW...)
	e.POST("/versions/:id/approve", h.ApproveVersion, protectedMW...)
	e.POST("/versions/:id/reject", h.RejectVersion, protectedMW...)
	e.POST("/versions/:id/publish", h.PublishVersion, protectedMW...)
	e.GET("/versions/:id/status", h.GetVersionStatus, standardMW...)
	e.GET("/versions/:id/approvals", h.GetVersionApprovals, standardMW...)
}

func actor(c echo.Context) (appver.Actor, error) {
	a, err := httpshared.ActorFromEcho(c)
	return appver.Actor{UserID: a.UserID}, err
}

func (h *Handler) CreateVersion(c echo.Context) error {
	documentID, err := uuid.Parse(c.Param("documentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid document ID"})
	}

	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Bind JSON request
	var req CreateVersionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	// Decode base64 content
	decodedContent, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "content must be valid base64 encoded"})
	}

	// Convert to io.Reader for application input
	contentReader := bytes.NewReader(decodedContent)

	// Call application service
	output, err := h.versions.Create(c.Request().Context(), appver.CreateInput{
		Actor:      a,
		DocumentID: documentID,
		Version:    req.Version,
		Content:    contentReader,
		Metadata:   req.Metadata,
	})
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusCreated, toCreateVersionResponse(output))
}
func (h *Handler) GetVersion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	view, err := h.versions.Get(c.Request().Context(), appver.GetInput{Actor: a, ID: id})
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, toVersionResponse(view))
}
func (h *Handler) ListVersionsByDocument(c echo.Context) error {
	documentID, err := uuid.Parse(c.Param("documentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid document ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	views, err := h.versions.ListByDocument(c.Request().Context(), appver.ListByDocumentInput{Actor: a, DocumentID: documentID})
	if err != nil {
		return handleError(c, err)
	}
	responses := make([]VersionResponse, 0, len(views))
	for _, v := range views {
		responses = append(responses, toVersionResponse(v))
	}
	return c.JSON(http.StatusOK, responses)
}
func (h *Handler) UpdateVersion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req UpdateVersionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	if err := h.versions.Update(c.Request().Context(), appver.UpdateInput{Actor: a, ID: id, Metadata: req.Metadata}); err != nil {
		return handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
func (h *Handler) SubmitVersion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	if err := h.versions.Submit(c.Request().Context(), appver.SubmitInput{Actor: a, ID: id}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *Handler) ReviewVersion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	if err := h.versions.Review(c.Request().Context(), appver.ReviewInput{Actor: a, ID: id}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *Handler) ApproveVersion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	var req ApproveVersionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}
	if err := h.versions.Approve(c.Request().Context(), appver.ApproveInput{Actor: a, ID: id, Role: req.Role, Comment: req.Comment}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *Handler) RejectVersion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	var req RejectVersionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}
	if err := h.versions.Reject(c.Request().Context(), appver.RejectInput{Actor: a, ID: id, Reason: req.Reason}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *Handler) PublishVersion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req PublishVersionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	if err := h.versions.Publish(c.Request().Context(), appver.PublishInput{
		Actor:       a,
		ID:          id,
		Destination: req.Destination,
	}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *Handler) GetVersionStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	view, err := h.versions.GetStatus(c.Request().Context(), appver.GetStatusInput{Actor: a, ID: id})
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, toVersionStatusResponse(view))
}
func (h *Handler) GetVersionApprovals(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}
	a, err := actor(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	view, err := h.versions.GetApprovals(c.Request().Context(), appver.GetApprovalsInput{Actor: a, ID: id})
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, toVersionApprovalsResponse(view))
}

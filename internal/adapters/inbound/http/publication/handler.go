package publication

import (
	"net/http"
	"strconv"

	appPublication "github.com/bbridges_11/document-registry/internal/application/publication"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for publications
type Handler struct {
	queries appPublication.QueryService
}

// NewHandler creates a new publication handler
func NewHandler(queries appPublication.QueryService) *Handler {
	return &Handler{
		queries: queries,
	}
}

// RegisterRoutes registers publication routes
// All routes require admin middleware
func (h *Handler) RegisterRoutes(e *echo.Group) {
	e.GET("/publications", h.ListAllPublications)
	e.GET("/publications/:id", h.GetPublication)
	e.GET("/versions/:version_id/publications", h.ListPublicationsByVersion)
	e.GET("/documents/:document_id/publications", h.ListPublicationsByDocument)
}

// GetPublication retrieves a single publication by ID
func (h *Handler) GetPublication(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid publication ID"})
	}

	view, err := h.queries.GetPublication(c.Request().Context(), appPublication.GetPublicationQuery{
		ID: id,
	})
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "publication not found"})
	}

	return c.JSON(http.StatusOK, toPublicationResponse(view))
}

// ListPublicationsByVersion retrieves all publications for a version
func (h *Handler) ListPublicationsByVersion(c echo.Context) error {
	versionID, err := uuid.Parse(c.Param("version_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid version ID"})
	}

	views, err := h.queries.ListPublicationsByVersion(c.Request().Context(), appPublication.ListPublicationsByVersionQuery{
		VersionID: versionID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve publications"})
	}

	responses := make([]PublicationResponse, len(views))
	for i, view := range views {
		responses[i] = toPublicationResponse(view)
	}

	return c.JSON(http.StatusOK, responses)
}

// ListPublicationsByDocument retrieves all publications for a document
func (h *Handler) ListPublicationsByDocument(c echo.Context) error {
	documentID, err := uuid.Parse(c.Param("document_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid document ID"})
	}

	views, err := h.queries.ListPublicationsByDocument(c.Request().Context(), appPublication.ListPublicationsByDocumentQuery{
		DocumentID: documentID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve publications"})
	}

	responses := make([]PublicationResponse, len(views))
	for i, view := range views {
		responses[i] = toPublicationResponse(view)
	}

	return c.JSON(http.StatusOK, responses)
}

// ListAllPublications retrieves all publications with pagination
func (h *Handler) ListAllPublications(c echo.Context) error {
	// Parse pagination parameters
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 50 // default
	}
	if limit > 100 {
		limit = 100 // max
	}

	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	result, err := h.queries.ListAllPublications(c.Request().Context(), appPublication.ListAllPublicationsQuery{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve publications"})
	}

	return c.JSON(http.StatusOK, toPublicationListResponse(result))
}

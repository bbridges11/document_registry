package document

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	httpshared "github.com/bbridges_11/document-registry/internal/adapters/inbound/http/shared"
	appdoc "github.com/bbridges_11/document-registry/internal/application/document"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct{ documents appdoc.UseCase }

func NewHandler(documents appdoc.UseCase) *Handler { return &Handler{documents: documents} }

type MiddlewareConfig struct {
	UserValidation echo.MiddlewareFunc
	RateLimit      echo.MiddlewareFunc
}

func (h *Handler) RegisterRoutes(e *echo.Echo, mw *MiddlewareConfig) {
	middlewares := []echo.MiddlewareFunc{}
	if mw != nil {
		if mw.UserValidation != nil {
			middlewares = append(middlewares, mw.UserValidation)
		}
		if mw.RateLimit != nil {
			middlewares = append(middlewares, mw.RateLimit)
		}
	}
	e.POST("/documents", h.CreateDocument, middlewares...)
	e.GET("/documents/search", h.SearchDocuments, middlewares...)
	e.GET("/documents/:id", h.GetDocument, middlewares...)
	e.GET("/documents", h.ListDocuments, middlewares...)
	e.PUT("/documents/:id", h.UpdateDocument, middlewares...)
}

func mustString(values map[string][]string, key string) string {
	if v := values[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

func parseIntParam(c echo.Context, key string, defaultValue int) int {
	val := c.QueryParam(key)
	if val == "" {
		return defaultValue
	}
	if intVal, err := strconv.Atoi(val); err == nil {
		return intVal
	}
	return defaultValue
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseTime(s string) (time.Time, error) {
	// Try ISO8601 format first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Try date-only format
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New(errors.CodeInvalidArgument, "invalid time format")
}

func parseMetadata(values map[string][]string) map[string]any {
	metadata := map[string]any{}
	if mv, ok := values["metadata"]; ok {
		for i := 0; i+1 < len(mv); i += 2 {
			metadata[mv[i]] = mv[i+1]
		}
	}
	return metadata
}
func fileFromForm(files []*multipart.FileHeader) (multipart.File, error) { return files[0].Open() }

func (h *Handler) CreateDocument(c echo.Context) error {
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid multipart form"})
	}
	files := form.File["content"]
	if len(files) == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "content file required"})
	}
	file, err := fileFromForm(files)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to open content file"})
	}
	defer file.Close()

	input := appdoc.CreateInput{Actor: appdoc.Actor{UserID: actor.UserID}, Name: mustString(form.Value, "name"), Description: mustString(form.Value, "description"), Tags: form.Value["tags"], DocumentType: mustString(form.Value, "document_type"), Version: mustString(form.Value, "version"), Content: file, Metadata: parseMetadata(form.Value)}
	out, err := h.documents.Create(c.Request().Context(), input)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusCreated, toCreateDocumentResponse(out))
}

func (h *Handler) GetDocument(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid document ID"})
	}
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	view, err := h.documents.Get(c.Request().Context(), appdoc.GetInput{Actor: appdoc.Actor{UserID: actor.UserID}, ID: id})
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, toDocumentResponse(view))
}

func (h *Handler) ListDocuments(c echo.Context) error {
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	views, err := h.documents.List(c.Request().Context(), appdoc.ListInput{Actor: appdoc.Actor{UserID: actor.UserID}, Limit: 50, Offset: 0})
	if err != nil {
		return handleError(c, err)
	}
	responses := make([]DocumentResponse, 0, len(views))
	for _, v := range views {
		responses = append(responses, toDocumentResponse(v))
	}
	return c.JSON(http.StatusOK, responses)
}

func (h *Handler) UpdateDocument(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid document ID"})
	}
	var req UpdateDocumentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	if err := h.documents.Update(c.Request().Context(), appdoc.UpdateInput{Actor: appdoc.Actor{UserID: actor.UserID}, ID: id, Name: req.Name, Description: req.Description, Tags: req.Tags}); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) SearchDocuments(c echo.Context) error {
	actor, err := httpshared.ActorFromEcho(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	// Parse search input from query parameters
	input := appdoc.SearchInput{
		Actor:        appdoc.Actor{UserID: actor.UserID},
		Query:        c.QueryParam("q"),
		Name:         c.QueryParam("name"),
		Description:  c.QueryParam("description"),
		DocumentType: c.QueryParam("document_type"),
		CreatedBy:    c.QueryParam("created_by"),
		SortField:    c.QueryParam("sort_field"),
		SortOrder:    c.QueryParam("sort_order"),
	}

	// Parse tags (comma-separated)
	if tags := c.QueryParam("tags"); tags != "" {
		input.Tags = parseCommaSeparated(tags)
	}

	// Parse date filters
	if createdAfter := c.QueryParam("created_after"); createdAfter != "" {
		if t, err := parseTime(createdAfter); err == nil {
			input.CreatedAfter = &t
		}
	}
	if createdBefore := c.QueryParam("created_before"); createdBefore != "" {
		if t, err := parseTime(createdBefore); err == nil {
			input.CreatedBefore = &t
		}
	}

	// Parse special filters
	input.MyDocuments = c.QueryParam("my_documents") == "true"
	input.MyStakeholderDocs = c.QueryParam("my_stakeholder_docs") == "true"
	input.PendingMyReview = c.QueryParam("pending_my_review") == "true"
	input.PendingMyApproval = c.QueryParam("pending_my_approval") == "true"
	input.RecentlyPublished = c.QueryParam("recently_published") == "true"

	// Parse pagination
	input.Page = parseIntParam(c, "page", 1)
	input.Size = parseIntParam(c, "size", 20)

	// Call service
	output, err := h.documents.Search(c.Request().Context(), input)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, toSearchResponse(output))
}

package user

import (
	"net/http"

	appUser "github.com/bbridges_11/document-registry/internal/application/user"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	commands appUser.CommandService
	queries  appUser.QueryService
}

func NewHandler(
	commands appUser.CommandService,
	queries appUser.QueryService,
) *Handler {
	return &Handler{
		commands: commands,
		queries:  queries,
	}
}

// MiddlewareConfig holds optional middleware for user routes
type MiddlewareConfig struct {
	UserValidation echo.MiddlewareFunc // Validate user exists
	AdminOnly      echo.MiddlewareFunc // Require admin role
}

// RegisterRoutes registers user routes with optional middleware
func (h *Handler) RegisterRoutes(e *echo.Echo, mw *MiddlewareConfig) {
	// Public routes (no middleware)
	e.POST("/users", h.CreateUser)
	e.GET("/users", h.ListUsers)

	// Protected routes (user validation only)
	userValidated := []echo.MiddlewareFunc{}
	if mw != nil && mw.UserValidation != nil {
		userValidated = append(userValidated, mw.UserValidation)
	}

	e.GET("/users/:id", h.GetUser, userValidated...)

	// Admin-only routes (user validation + admin check)
	adminOnly := []echo.MiddlewareFunc{}
	if mw != nil {
		if mw.UserValidation != nil {
			adminOnly = append(adminOnly, mw.UserValidation)
		}
		if mw.AdminOnly != nil {
			adminOnly = append(adminOnly, mw.AdminOnly)
		}
	}

	e.PUT("/users/:id", h.UpdateUser, adminOnly...)
	e.DELETE("/users/:id", h.DeleteUser, adminOnly...)
	e.POST("/users/:id/deactivate", h.DeactivateUser, adminOnly...)
	e.POST("/users/:id/activate", h.ActivateUser, adminOnly...)
}

// CreateUser handles POST /users
func (h *Handler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	cmd := appUser.CreateUserCommand{
		ExternalID: req.ExternalID,
		Email:      req.Email,
		Name:       req.Name,
		Role:       req.Role,
	}

	dto, err := h.commands.CreateUser(c.Request().Context(), cmd)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusCreated, ToUserResponse(dto))
}

// GetUser handles GET /users/:id
func (h *Handler) GetUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
	}

	query := appUser.GetUserQuery{ID: id}
	dto, err := h.queries.GetUser(c.Request().Context(), query)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, ToUserResponse(dto))
}

// ListUsers handles GET /users
func (h *Handler) ListUsers(c echo.Context) error {
	query := appUser.ListUsersQuery{}
	dtos, err := h.queries.ListUsers(c.Request().Context(), query)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, ToUsersResponse(dtos))
}

// UpdateUser handles PUT /users/:id
func (h *Handler) UpdateUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
	}

	cmd := appUser.UpdateUserCommand{
		ID:   id,
		Name: req.Name,
		Role: req.Role,
	}

	dto, err := h.commands.UpdateUser(c.Request().Context(), cmd)
	if err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, ToUserResponse(dto))
}

// DeleteUser handles DELETE /users/:id
func (h *Handler) DeleteUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
	}

	if err := h.commands.DeleteUser(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// DeactivateUser handles POST /users/:id/deactivate
func (h *Handler) DeactivateUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
	}

	if err := h.commands.DeactivateUser(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user deactivated"})
}

// ActivateUser handles POST /users/:id/activate
func (h *Handler) ActivateUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
	}

	if err := h.commands.ActivateUser(c.Request().Context(), id); err != nil {
		return handleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user activated"})
}

// handleError converts application errors to HTTP responses
func handleError(c echo.Context, err error) error {
	// You can import your errors package and do proper error code mapping
	// For now, simple error handling
	if err.Error() == "not found" {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	}
	if err.Error() == "user with this email already exists" {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
}

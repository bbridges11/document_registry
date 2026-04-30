package subscription

import (
	"net/http"

	appSubscription "github.com/bbridges_11/document-registry/internal/application/subscription"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for subscriptions
type Handler struct {
	service appSubscription.Service
}

// NewHandler creates a new subscription handler
func NewHandler(service appSubscription.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers subscription routes
// All routes require admin middleware (applied in bootstrap)
func (h *Handler) RegisterRoutes(e *echo.Group) {
	e.POST("/subscriptions/subscribe/:user_id", h.Subscribe)
	e.POST("/subscriptions/unsubscribe/:user_id", h.Unsubscribe)
	e.GET("/subscriptions", h.ListSubscriptions)
}

// Subscribe subscribes a user's email to notifications
func (h *Handler) Subscribe(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
	}

	err = h.service.Subscribe(c.Request().Context(), appSubscription.SubscribeInput{
		UserID: userID,
	})
	if err != nil {
		// Check error type
		if err.Error() == "user not found" || err.Error() == "user not found or inactive" {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		}
		if err.Error() == "user has no email address" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user has no email address"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to subscribe user"})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "User subscribed successfully. Confirmation email sent.",
	})
}

// Unsubscribe removes a user's email subscription
func (h *Handler) Unsubscribe(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
	}

	err = h.service.Unsubscribe(c.Request().Context(), appSubscription.UnsubscribeInput{
		UserID: userID,
	})
	if err != nil {
		// Check error type
		if err.Error() == "user not found" || err.Error() == "user not found or inactive" {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		}
		if err.Error() == "user has no email address" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user has no email address"})
		}
		if err.Error() == "subscription not found for email: "+c.Param("user_id") {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user is not subscribed"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to unsubscribe user"})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "User unsubscribed successfully.",
	})
}

// ListSubscriptions returns all current email subscriptions
func (h *Handler) ListSubscriptions(c echo.Context) error {
	views, err := h.service.ListSubscriptions(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list subscriptions"})
	}

	responses := make([]SubscriptionResponse, len(views))
	for i, view := range views {
		responses[i] = SubscriptionResponse{
			Email:  view.Email,
			Status: view.Status,
		}
	}

	return c.JSON(http.StatusOK, responses)
}

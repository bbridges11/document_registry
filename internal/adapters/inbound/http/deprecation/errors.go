package deprecation

import (
	"net/http"

	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/labstack/echo/v4"
)

// handleError maps application errors to HTTP status codes
func handleError(c echo.Context, err error) error {
	// General error codes
	if errors.IsCode(err, errors.CodeNotFound) {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeInvalidArgument) {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeConflict) {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeUnauthorized) {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeForbidden) {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodePrecondition) {
		return c.JSON(http.StatusPreconditionFailed, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeInvalidState) {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	// Deprecation-specific error codes
	if errors.IsCode(err, errors.CodeVersionNotPublished) {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeVersionAlreadyDeprecated) {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeVersionDeprecating) {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeMultiplePublished) {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeCannotDeprecate) {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeDeprecationNotFound) {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	}
	if errors.IsCode(err, errors.CodeDeprecationNotPending) {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	// Internal error
	return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
}

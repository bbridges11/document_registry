package validation

import (
	"net/http"

	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/labstack/echo/v4"
)

func handleError(c echo.Context, err error) error {
	if errors.IsCode(err, errors.CodeValidationFailed) {
		return c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
	}

	if errors.IsCode(err, errors.CodeInvalidArgument) {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	if errors.IsCode(err, errors.CodeUnauthorized) {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	if errors.IsCode(err, errors.CodeForbidden) {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
	}

	// Internal server error
	return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
}

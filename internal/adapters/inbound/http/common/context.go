package common

import (
	"context"

	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/labstack/echo/v4"
)

// GetActor extracts the Actor from the echo context
// This should only be called in handlers after UserValidationMiddleware has run
func GetActor(c echo.Context) (Actor, error) {
	actor, ok := c.Request().Context().Value(actorContextKey).(Actor)
	if !ok {
		return Actor{}, errors.New(errors.CodeInternal, "actor not found in context")
	}
	return actor, nil
}

// WithActor adds an Actor to a context
// Used by middleware to enrich the request context
func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey, actor)
}

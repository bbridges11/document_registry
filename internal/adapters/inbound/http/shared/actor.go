package shared

import (
	"errors"

	"github.com/labstack/echo/v4"
)

type Actor struct {
	UserID string
}

func ActorFromEcho(c echo.Context) (Actor, error) {
	userID := c.Request().Header.Get("X-User-ID")
	if userID == "" {
		return Actor{}, errors.New("missing user identity")
	}
	return Actor{UserID: userID}, nil
}

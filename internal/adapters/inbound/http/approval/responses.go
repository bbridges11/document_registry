package approval

import (
	"time"

	"github.com/google/uuid"
)

type ApprovalResponse struct {
	ID        uuid.UUID `json:"id"`
	VersionID uuid.UUID `json:"version_id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Role      string    `json:"role"`
	Approved  bool      `json:"approved"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}
type ApprovalSummaryResponse struct {
	VersionID uuid.UUID      `json:"version_id"`
	Required  map[string]int `json:"required"`
	Received  map[string]int `json:"received"`
	Remaining map[string]int `json:"remaining"`
	Complete  bool           `json:"complete"`
}
type ErrorResponse struct {
	Error string `json:"error"`
}

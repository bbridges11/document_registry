package stakeholder

import (
	"time"

	"github.com/google/uuid"
)

type StakeholderResponse struct {
	DocumentID uuid.UUID `json:"document_id"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
}
type ErrorResponse struct {
	Error string `json:"error"`
}

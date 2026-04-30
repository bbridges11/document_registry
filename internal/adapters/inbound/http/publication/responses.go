package publication

import "time"

// PublicationResponse represents a publication in HTTP responses
type PublicationResponse struct {
	ID           string    `json:"id"`
	VersionID    string    `json:"version_id"`
	DocumentID   string    `json:"document_id"`
	PublishedTo  string    `json:"published_to"`
	PublishedBy  string    `json:"published_by"`
	PublishedAt  time.Time `json:"published_at"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// PublicationListResponse represents a paginated list of publications
type PublicationListResponse struct {
	Publications []PublicationResponse `json:"publications"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

// ErrorResponse represents an error in HTTP responses
type ErrorResponse struct {
	Error string `json:"error"`
}

package subscription

// SubscriptionResponse represents a subscription in HTTP responses
type SubscriptionResponse struct {
	Email  string `json:"email"`
	Status string `json:"status"`
}

// MessageResponse represents a success message
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error
type ErrorResponse struct {
	Error string `json:"error"`
}

package deprecation

// RequestDeprecationRequest represents the request to deprecate a version
type RequestDeprecationRequest struct {
	Reason          string `json:"reason"`
	DeprecationNote string `json:"deprecation_note"`
}

// ApproveDeprecationRequest represents the request to approve a deprecation
type ApproveDeprecationRequest struct {
	Role    string `json:"role"`
	Comment string `json:"comment"`
}

// RejectDeprecationRequest represents the request to reject a deprecation
type RejectDeprecationRequest struct {
	Reason string `json:"reason"`
}

package deprecation

// DeprecationStatus represents the state of a deprecation request
type DeprecationStatus string

const (
	// DeprecationStatusPending indicates deprecation is awaiting approval
	DeprecationStatusPending DeprecationStatus = "PENDING"

	// DeprecationStatusApproved indicates deprecation has been approved
	DeprecationStatusApproved DeprecationStatus = "APPROVED"

	// DeprecationStatusRejected indicates deprecation was rejected
	DeprecationStatusRejected DeprecationStatus = "REJECTED"

	// DeprecationStatusCanceled indicates deprecation was canceled by requestor
	DeprecationStatusCanceled DeprecationStatus = "CANCELED"
)

// String returns the string representation of the status
func (s DeprecationStatus) String() string {
	return string(s)
}

// IsPending returns true if deprecation is pending approval
func (s DeprecationStatus) IsPending() bool {
	return s == DeprecationStatusPending
}

// IsApproved returns true if deprecation has been approved
func (s DeprecationStatus) IsApproved() bool {
	return s == DeprecationStatusApproved
}

// IsRejected returns true if deprecation was rejected
func (s DeprecationStatus) IsRejected() bool {
	return s == DeprecationStatusRejected
}

// IsCanceled returns true if deprecation was canceled
func (s DeprecationStatus) IsCanceled() bool {
	return s == DeprecationStatusCanceled
}

// IsTerminal returns true if deprecation is in a final state
func (s DeprecationStatus) IsTerminal() bool {
	return s == DeprecationStatusApproved || s == DeprecationStatusRejected || s == DeprecationStatusCanceled
}

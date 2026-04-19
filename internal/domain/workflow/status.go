package workflow

type Status string

const (
	StatusDraft       Status = "DRAFT"
	StatusSubmitted   Status = "SUBMITTED"
	StatusInReview    Status = "IN_REVIEW"
	StatusApproved    Status = "APPROVED"
	StatusRejected    Status = "REJECTED"
	StatusPublished   Status = "PUBLISHED"
	StatusDeprecating Status = "DEPRECATING" // NEW: Version is locked pending deprecation approval
	StatusDeprecated  Status = "DEPRECATED"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsEditable() bool {
	return s == StatusDraft
}

func (s Status) IsTerminal() bool {
	return s == StatusRejected || s == StatusDeprecated
}

func (s Status) IsPublishable() bool {
	return s == StatusApproved
}

func (s Status) IsDeprecated() bool {
	return s == StatusDeprecated
}

func (s Status) IsDeprecating() bool {
	return s == StatusDeprecating
}

func (s Status) CanBeDeprecated() bool {
	return s == StatusPublished
}

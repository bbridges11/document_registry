package workflow

// Type represents the type of workflow a document follows
type Type string

const (
	// TypeApprovalBased - Document requires approval before publication
	// Flow: Draft → Submitted → Under Review → Approved/Rejected → Published
	TypeApprovalBased Type = "approval_based"

	// TypeAutoPublish - Document auto-publishes without approval (future)
	// Flow: Draft → Published
	TypeAutoPublish Type = "auto_publish"

	// TypeStrictApproval - Document requires multiple levels of approval (future)
	// Flow: Draft → Submitted → Review L1 → Review L2 → Approved/Rejected → Published
	TypeStrictApproval Type = "strict_approval"
)

// String returns the string representation
func (t Type) String() string {
	return string(t)
}

package document

// DeprecationPolicy defines how deprecation works for a document type
type DeprecationPolicy struct {
	// ManualRequiresApproval - Does manual deprecation request need approval?
	ManualRequiresApproval bool

	// AutoRequiresApproval - Does auto-deprecation (superseded) need approval?
	AutoRequiresApproval bool
}

// Common deprecation policies
var (
	// DeprecationPolicyStandard - Manual needs approval, auto doesn't
	DeprecationPolicyStandard = DeprecationPolicy{
		ManualRequiresApproval: true,
		AutoRequiresApproval:   false,
	}

	// DeprecationPolicyStrict - Both manual and auto need approval
	DeprecationPolicyStrict = DeprecationPolicy{
		ManualRequiresApproval: true,
		AutoRequiresApproval:   true,
	}

	// DeprecationPolicyNone - No approvals needed (auto-approve everything)
	DeprecationPolicyNone = DeprecationPolicy{
		ManualRequiresApproval: false,
		AutoRequiresApproval:   false,
	}
)

package approval

// PolicyType identifies which approval policy to use
type PolicyType string

const (
	// PolicyPattern - Standard pattern approval policy
	// Requires stakeholder approvals based on pattern scope
	PolicyPattern PolicyType = "pattern"

	// PolicyNone - No approval required (future)
	// Used for document types that auto-publish
	PolicyNone PolicyType = "none"

	// PolicyStrict - Strict approval policy (future)
	// Requires multiple approval levels (e.g., legal + executive)
	PolicyStrict PolicyType = "strict"
)

// String returns the string representation
func (p PolicyType) String() string {
	return string(p)
}

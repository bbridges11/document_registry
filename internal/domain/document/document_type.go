package document

import (
	"fmt"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	"github.com/bbridges_11/document-registry/pkg/errors"
)

// DocumentType represents a type of document in the registry
// This is a value object that encapsulates all type-specific behavior
type DocumentType struct {
	code               string
	displayName        string
	workflowType       workflow.Type
	approvalPolicy     *approval.PolicyType
	deprecationPolicy  *DeprecationPolicy
	validationRequired bool
}

// Code returns the document type code (used for persistence and API)
func (dt DocumentType) Code() string {
	return dt.code
}

// DisplayName returns the human-readable name
func (dt DocumentType) DisplayName() string {
	return dt.displayName
}

// WorkflowType returns the workflow type this document follows
func (dt DocumentType) WorkflowType() workflow.Type {
	return dt.workflowType
}

// RequiresApproval returns true if this document type requires approval
func (dt DocumentType) RequiresApproval() bool {
	return dt.approvalPolicy != nil
}

// ApprovalPolicy returns the approval policy type
// Panics if RequiresApproval() is false - caller must check first
func (dt DocumentType) ApprovalPolicy() approval.PolicyType {
	if dt.approvalPolicy == nil {
		panic(fmt.Sprintf("document type %s does not require approval", dt.code))
	}
	return *dt.approvalPolicy
}

// SupportsDeprecation returns true if this document type supports deprecation
func (dt DocumentType) SupportsDeprecation() bool {
	return dt.deprecationPolicy != nil
}

// DeprecationPolicy returns the deprecation policy
// Panics if SupportsDeprecation() is false - caller must check first
func (dt DocumentType) DeprecationPolicy() DeprecationPolicy {
	if dt.deprecationPolicy == nil {
		panic(fmt.Sprintf("document type %s does not support deprecation", dt.code))
	}
	return *dt.deprecationPolicy
}

// RequiresValidation returns true if content must be validated
func (dt DocumentType) RequiresValidation() bool {
	return dt.validationRequired
}

// Equals compares two document types
func (dt DocumentType) Equals(other DocumentType) bool {
	return dt.code == other.code
}

// String returns the code representation
func (dt DocumentType) String() string {
	return dt.code
}

// Registry of all document types
var (
	patternApprovalPolicy    = approval.PolicyPattern
	patternDeprecationPolicy = DeprecationPolicyStandard

	// TypePattern - Architectural/design patterns requiring review and approval
	TypePattern = DocumentType{
		code:               "pattern",
		displayName:        "Pattern",
		workflowType:       workflow.TypeApprovalBased,
		approvalPolicy:     &patternApprovalPolicy,
		deprecationPolicy:  &patternDeprecationPolicy,
		validationRequired: true,
	}
)

// Global registry mapping code to DocumentType
var typeRegistry = map[string]DocumentType{
	"pattern": TypePattern,
}

// ParseDocumentType parses a document type code string into a DocumentType value object
func ParseDocumentType(code string) (DocumentType, error) {
	dt, ok := typeRegistry[code]
	if !ok {
		return DocumentType{}, errors.New(
			errors.CodeInvalidArgument,
			fmt.Sprintf("unknown document type: %s", code),
		)
	}
	return dt, nil
}

// MustParseDocumentType parses a document type code or panics
// Use only when the code is guaranteed to be valid (e.g., from database)
func MustParseDocumentType(code string) DocumentType {
	dt, err := ParseDocumentType(code)
	if err != nil {
		panic(fmt.Sprintf("invalid document type code: %s", code))
	}
	return dt
}

// AllDocumentTypes returns all registered document types
func AllDocumentTypes() []DocumentType {
	return []DocumentType{
		TypePattern,
	}
}

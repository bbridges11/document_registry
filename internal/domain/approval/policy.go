package approval

import "github.com/bbridges_11/document-registry/pkg/errors"

type Policy interface {
	RequiredApprovals() map[ApprovalRole]int
	IsSatisfied(approvals []*Approval) bool
	ValidateApproval(approval *Approval, authorUserID string, userRoles []ApprovalRole, approvals []*Approval) error
}

// PatternPolicy defines approval requirements for "pattern" document type
// Requires: 1 technical, 1 architect, 1 product
type PatternPolicy struct{}

func NewPatternPolicy() *PatternPolicy {
	return &PatternPolicy{}
}

func (p *PatternPolicy) RequiredApprovals() map[ApprovalRole]int {
	return map[ApprovalRole]int{
		ApprovalRoleTechnical: 1,
		ApprovalRoleArchitect: 1,
		ApprovalRoleProduct:   1,
	}
}

func (p *PatternPolicy) IsSatisfied(approvals []*Approval) bool {
	counts := make(map[ApprovalRole]int)
	userRoles := make(map[string]map[ApprovalRole]bool)

	for _, approval := range approvals {
		if !approval.Approved() {
			continue
		}

		// Track which roles this user has already approved with
		if userRoles[approval.UserID()] == nil {
			userRoles[approval.UserID()] = make(map[ApprovalRole]bool)
		}

		// Same user cannot satisfy multiple roles
		if len(userRoles[approval.UserID()]) > 0 {
			continue
		}

		userRoles[approval.UserID()][approval.Role()] = true
		counts[approval.Role()]++
	}

	required := p.RequiredApprovals()
	for role, count := range required {
		if counts[role] < count {
			return false
		}
	}

	return true
}

func (p *PatternPolicy) ValidateApproval(approval *Approval, authorUserID string, userRoles []ApprovalRole, approvals []*Approval) error {
	// Author cannot approve their own version
	if approval.UserID() == authorUserID {
		return errors.New(errors.CodeForbidden, "author cannot approve own version")
	}

	// User must have the role they're claiming
	hasRole := false
	for _, role := range userRoles {
		if role == approval.Role() {
			hasRole = true
			break
		}
	}
	if !hasRole {
		return errors.New(errors.CodeForbidden, "user does not have required approval role")
	}

	// Same user cannot satisfy multiple roles
	for _, existingApproval := range approvals {
		if existingApproval.UserID() == approval.UserID() && existingApproval.Approved() {
			return errors.New(errors.CodeConflict, "user has already approved with a different role")
		}
	}

	return nil
}

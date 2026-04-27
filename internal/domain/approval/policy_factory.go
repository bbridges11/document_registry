package approval

import (
	"fmt"

	"github.com/bbridges_11/document-registry/pkg/errors"
)

type PolicyFactory struct{}

func NewPolicyFactory() *PolicyFactory {
	return &PolicyFactory{}
}

func (f *PolicyFactory) GetPolicy(policyType PolicyType) (Policy, error) {
	switch policyType {
	case PolicyPattern:
		return NewPatternPolicy(), nil
	default:
		return nil, errors.New(errors.CodeInvalidArgument, fmt.Sprintf("unknown approval policy type: %s", policyType))
	}
}

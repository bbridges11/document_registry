package approval

import (
	"fmt"

	"github.com/bbridges_11/document-registry/pkg/errors"
)

type PolicyFactory struct{}

func NewPolicyFactory() *PolicyFactory {
	return &PolicyFactory{}
}

func (f *PolicyFactory) GetPolicy(documentType string) (Policy, error) {
	switch documentType {
	case "pattern":
		return NewPatternPolicy(), nil
	default:
		return nil, errors.New(errors.CodeInvalidArgument, fmt.Sprintf("unknown document type: %s", documentType))
	}
}

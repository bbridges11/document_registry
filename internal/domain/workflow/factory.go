package workflow

import (
	"fmt"

	"github.com/bbridges_11/document-registry/pkg/errors"
)

type Workflow interface {
	CanTransition(from Status, action Action) error
	NextStatus(current Status, action Action) (Status, error)
	AllowedActions(status Status) []Action
}

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) GetWorkflow(documentType string) (Workflow, error) {
	switch documentType {
	case "pattern":
		return NewPatternWorkflow(), nil
	default:
		return nil, errors.New(errors.CodeInvalidArgument, fmt.Sprintf("unknown document type: %s", documentType))
	}
}

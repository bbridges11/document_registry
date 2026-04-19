package workflow

import (
	"fmt"

	"github.com/bbridges_11/document-registry/pkg/errors"
)

type Action string

const (
	ActionSubmit  Action = "submit"
	ActionReview  Action = "review"
	ActionApprove Action = "approve"
	ActionReject  Action = "reject"
	ActionPublish Action = "publish"
)

// PatternWorkflow implements the workflow for "pattern" document type
// DRAFT → SUBMITTED → IN_REVIEW → APPROVED → PUBLISHED
//
//	   ↓
//	REJECTED
type PatternWorkflow struct{}

func NewPatternWorkflow() *PatternWorkflow {
	return &PatternWorkflow{}
}

func (w *PatternWorkflow) CanTransition(from Status, action Action) error {
	switch from {
	case StatusDraft:
		if action != ActionSubmit {
			return errors.New(errors.CodeInvalidState, fmt.Sprintf("invalid action %s from DRAFT", action))
		}
		return nil

	case StatusSubmitted:
		if action != ActionReview {
			return errors.New(errors.CodeInvalidState, fmt.Sprintf("invalid action %s from SUBMITTED", action))
		}
		return nil

	case StatusInReview:
		if action != ActionApprove && action != ActionReject {
			return errors.New(errors.CodeInvalidState, fmt.Sprintf("invalid action %s from IN_REVIEW", action))
		}
		return nil

	case StatusApproved:
		if action != ActionPublish {
			return errors.New(errors.CodeInvalidState, fmt.Sprintf("invalid action %s from APPROVED", action))
		}
		return nil

	case StatusRejected:
		return errors.New(errors.CodeInvalidState, "cannot transition from terminal REJECTED state")

	case StatusPublished:
		return errors.New(errors.CodeInvalidState, "cannot transition from PUBLISHED state")

	case StatusDeprecated:
		return errors.New(errors.CodeInvalidState, "cannot transition from DEPRECATED state")

	default:
		return errors.New(errors.CodeInvalidState, fmt.Sprintf("unknown status: %s", from))
	}
}

func (w *PatternWorkflow) NextStatus(current Status, action Action) (Status, error) {
	if err := w.CanTransition(current, action); err != nil {
		return "", err
	}

	switch action {
	case ActionSubmit:
		return StatusSubmitted, nil
	case ActionReview:
		return StatusInReview, nil
	case ActionApprove:
		return StatusApproved, nil
	case ActionReject:
		return StatusRejected, nil
	case ActionPublish:
		return StatusPublished, nil
	default:
		return "", errors.New(errors.CodeInvalidState, fmt.Sprintf("unknown action: %s", action))
	}
}

func (w *PatternWorkflow) AllowedActions(status Status) []Action {
	switch status {
	case StatusDraft:
		return []Action{ActionSubmit}
	case StatusSubmitted:
		return []Action{ActionReview}
	case StatusInReview:
		return []Action{ActionApprove, ActionReject}
	case StatusApproved:
		return []Action{ActionPublish}
	default:
		return []Action{}
	}
}

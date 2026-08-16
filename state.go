package editorial

import (
	"github.com/tinywasm/fmt"
)

type State uint8

const (
	StateDraft State = iota // 0
	StateInReview
	StateChangesRequested
	StateApproved
	StatePublished
	StateRetired
)

func (s State) String() string {
	switch s {
	case StateDraft:
		return "draft"
	case StateInReview:
		return "in_review"
	case StateChangesRequested:
		return "changes_requested"
	case StateApproved:
		return "approved"
	case StatePublished:
		return "published"
	case StateRetired:
		return "retired"
	default:
		return "unknown"
	}
}

var ErrInvalidTransition = fmt.Err("invalid state transition")

type transitionRule struct {
	From State
	To   State
}

var allowedTransitions = []transitionRule{
	{StateDraft, StateInReview},
	{StateInReview, StateApproved},
	{StateInReview, StateChangesRequested},
	{StateChangesRequested, StateInReview},
	{StateApproved, StatePublished},
	{StateApproved, StateChangesRequested},
	{StatePublished, StateRetired},
}

func IsValidTransition(from, to State) bool {
	for i := 0; i < len(allowedTransitions); i++ {
		if allowedTransitions[i].From == from && allowedTransitions[i].To == to {
			return true
		}
	}
	return false
}

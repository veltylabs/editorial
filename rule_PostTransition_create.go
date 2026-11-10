package editorial

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/time"
)

func (t *PostTransition) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	var existing Post
	if err := tx.ReadOne(&existing, t.PostId); err != nil {
		return 0, err
	}

	currentState := State(existing.State)
	targetState := State(t.ToState)

	if currentState != StateInReview {
		return 0, &model.FieldError{Field: "to_state", Err: ErrInvalidTransition}
	}

	switch targetState {
	case StateApproved:
		// allowed
	case StateChangesRequested:
		if t.Reason == "" {
			return 0, &model.FieldError{Field: "reason", Err: ErrReasonRequired}
		}
	default:
		return 0, &model.FieldError{Field: "to_state", Err: ErrInvalidTransition}
	}

	t.ActorId = a.UserID
	t.FromState = int64(currentState)
	t.At = time.NowNano()

	existing.State = t.ToState
	existing.UpdatedAt = t.At
	if err := tx.UpdateFields(&existing, existing.Id, []string{"state", "updated_at"}); err != nil {
		return 0, err
	}

	return crud.ApplyAndPublish, nil
}

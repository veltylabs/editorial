package editorial

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/time"
)

func (next *Post) Update(a crud.CurrentUser, tx crud.Tx, old *Post, fields []string, m *Module) (crud.RuleAnswer, error) {
	hasState := false
	for _, f := range fields {
		if f == "state" {
			hasState = true
			break
		}
	}

	if hasState {
		targetState := State(next.State)
		currentState := State(old.State)

		switch targetState {
		case StateInReview:
			if currentState != StateDraft {
				return 0, &model.FieldError{Field: "state", Err: ErrInvalidTransition}
			}
		case StateRetired:
			// "* -> Retired" is allowed
		default:
			return 0, &model.FieldError{Field: "state", Err: ErrInvalidTransition}
		}

		trans := &PostTransition{
			PostId:    old.Id,
			FromState: int64(currentState),
			ToState:   int64(targetState),
			ActorId:   a.UserID,
			At:        time.NowNano(),
		}
		if err := tx.Create(trans); err != nil {
			return 0, err
		}
	}
	return crud.ApplyAndPublish, nil
}

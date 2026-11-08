package editorial

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/time"
	"webtyp.com/events"
	"webtyp.com/fmt"
)

func (m *Module) CreatePost(c crud.Caller, list PostList) error {
	now := time.NowNano()
	for _, p := range list {
		if p.Id == "" {
			p.Id = m.deps.IDs.NewID()
		}
		p.TenantId = m.tenantID
		p.State = int64(StateDraft)
		p.CreatedAt = now
		p.UpdatedAt = now
		if err := p.Validate(model.ActionCreate); err != nil {
			return err
		}
	}
	return m.db.Tx(func(tx *orm.DB) error {
		for _, p := range list {
			if err := tx.Create(p); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *Module) UpdatePost(c crud.Caller, ids, fields []string, rec Post) error {
	// Only handling state transitions Draft->InReview and *->Retired for Author
	hasState := false
	for _, f := range fields {
		if f == "state" {
			hasState = true
			break
		}
	}

	return m.db.Tx(func(tx *orm.DB) error {
		for _, id := range ids {
			existing, err := m.getPost(id)
			if err != nil {
				return err
			}

			rec.Id = id
			rec.TenantId = m.tenantID
			rec.CreatedAt = existing.CreatedAt
			rec.PublishedAt = existing.PublishedAt
			rec.UpdatedAt = time.NowNano()

			if hasState {
				targetState := State(rec.State)
				currentState := State(existing.State)

				if targetState == StateInReview {
					if currentState != StateDraft {
						return &model.FieldError{Field: "state", Err: ErrInvalidTransition}
					}
				} else if targetState == StateRetired {
					// allowed from anywhere based on instructions "* -> Retired"
				} else {
					return &model.FieldError{Field: "state", Err: ErrInvalidTransition}
				}

				// Write transition
				trans := &PostTransition{
					Id:        m.deps.IDs.NewID(),
					TenantId:  m.tenantID,
					PostId:    id,
					FromState: int64(currentState),
					ToState:   int64(targetState),
					ActorId:   c.UserID,
					At:        time.NowNano(),
				}
				if err := trans.Validate(model.ActionCreate); err != nil {
					return err
				}
				if err := tx.Create(trans); err != nil {
					return err
				}
				if m.deps.Publisher != nil {
					m.deps.Publisher.Publish(events.Event{
						Topic:   EventPostTransitioned,
						Payload: trans,
					})
				}
			} else {
				// Regular update - don't mutate state
				rec.State = existing.State
			}

			if err := rec.Validate(model.ActionUpdate); err != nil {
				return err
			}

			if err := tx.Update(&rec, orm.Eq("tenant_id", m.tenantID), orm.Eq("id", id)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *Module) DeletePost(c crud.Caller, ids []string) error {
	return m.db.Tx(func(tx *orm.DB) error {
		for _, id := range ids {
			existing, err := m.getPost(id)
			if err != nil {
				return err
			}
			if err := tx.Delete(&Post{}, orm.Eq("tenant_id", m.tenantID), orm.Eq("id", existing.Id)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *Module) CreatePostTransition(c crud.Caller, list PostTransitionList) error {
	return m.db.Tx(func(tx *orm.DB) error {
		for _, t := range list {
			existing, err := m.getPost(t.PostId)
			if err != nil {
				return err
			}

			currentState := State(existing.State)
			targetState := State(t.ToState)

			if currentState != StateInReview {
				return &model.FieldError{Field: "to_state", Err: ErrInvalidTransition}
			}

			if targetState == StateApproved {
				// allowed
			} else if targetState == StateChangesRequested {
				if t.Reason == "" {
					return &model.FieldError{Field: "reason", Err: ErrReasonRequired}
				}
			} else {
				return &model.FieldError{Field: "to_state", Err: ErrInvalidTransition}
			}

			t.Id = m.deps.IDs.NewID()
			t.TenantId = m.tenantID
			t.ActorId = c.UserID
			t.FromState = int64(currentState)
			t.At = time.NowNano()

			if err := t.Validate(model.ActionCreate); err != nil {
				return err
			}
			if err := tx.Create(t); err != nil {
				return err
			}

			// Update Post
			existing.State = t.ToState
			existing.UpdatedAt = t.At
			if err := existing.Validate(model.ActionUpdate); err != nil {
				return err
			}
			if err := tx.Update(existing, orm.Eq("tenant_id", m.tenantID), orm.Eq("id", existing.Id)); err != nil {
				return err
			}

			if m.deps.Publisher != nil {
				m.deps.Publisher.Publish(events.Event{
					Topic:   EventPostTransitioned,
					Payload: t,
				})
			}
		}
		return nil
	})
}

func (m *Module) UpdatePostTransition(c crud.Caller, ids, fields []string, rec PostTransition) error {
	return fmt.Err("editorial: not supported")
}

func (m *Module) DeletePostTransition(c crud.Caller, ids []string) error {
	return fmt.Err("editorial: not supported")
}

func (m *Module) CreatePublication(c crud.Caller, list PublicationList) error {
	// Typically MarkPublished handles this, but if exposed via generic create...
	// We'll leave the base implementation. Note: the plan says "MountPublicationRead" only.
	// So we might not need this mount, but crud methods are good to have.
	return m.db.Tx(func(tx *orm.DB) error {
		for _, p := range list {
			p.TenantId = m.tenantID
			if err := p.Validate(model.ActionCreate); err != nil {
				return err
			}
			if err := tx.Create(p); err != nil {
				return err
			}
			if m.deps.Publisher != nil {
				m.deps.Publisher.Publish(events.Event{
					Topic:   EventPostPublished,
					Payload: p,
				})
			}
		}
		return nil
	})
}

func (m *Module) UpdatePublication(c crud.Caller, ids, fields []string, rec Publication) error {
	return fmt.Err("editorial: not supported")
}

func (m *Module) DeletePublication(c crud.Caller, ids []string) error {
	return fmt.Err("editorial: not supported")
}

package editorial

import (
	"github.com/tinywasm/ddl"
	"github.com/tinywasm/events"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage"
	"github.com/tinywasm/time"
)

var (
	ErrNotFound       = fmt.Err("editorial: not found")
	ErrAlreadyExists  = fmt.Err("editorial: already exists")
	ErrReasonRequired = fmt.Err("editorial: reason is required when requesting changes")
)

const (
	EventPostTransitioned = "editorial.post.transitioned"
	EventPostPublished    = "editorial.post.published"
)

type Deps struct {
	IDs       model.IDGenerator
	Publisher events.Publisher
}

type Module struct {
	db   *orm.DB
	deps Deps
}

func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("editorial: IDs generator is required")
	}
	m := &Module{
		db:   db,
		deps: deps,
	}

	if ddlCompiler, ok := db.RawConn().(ddl.Compiler); ok {
		d := ddl.New(db.RawConn(), ddlCompiler)
		if err := d.CreateTable(&Post{}); err != nil {
			return nil, err
		}
		if err := d.CreateTable(&PostTransition{}); err != nil {
			return nil, err
		}
		if err := d.CreateTable(&Publication{}); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *Module) Submit(tenantId, postId, actorId string) error {
	return m.transitionPost(tenantId, postId, actorId, StateInReview, "")
}

func (m *Module) Approve(tenantId, postId, actorId string) error {
	return m.transitionPost(tenantId, postId, actorId, StateApproved, "")
}

func (m *Module) RequestChanges(tenantId, postId, actorId, reason string) error {
	trimmed := reason
	if trimmed == "" {
		return ErrReasonRequired
	}
	return m.transitionPost(tenantId, postId, actorId, StateChangesRequested, reason)
}

func (m *Module) MarkPublished(tenantId, postId, channel, externalRef string) error {
	if channel == "" {
		channel = ChannelWeb
	}

	var existingPubs PublicationList
	err := m.db.Query(&Publication{}).
		Where("tenant_id").Eq(tenantId).
		Where("post_id").Eq(postId).
		Where("channel").Eq(channel).
		ReadAll(
			func() model.Model { return &Publication{} },
			func(rec model.Model) { existingPubs = append(existingPubs, rec.(*Publication)) },
		)
	if err != nil && err != orm.ErrNotFound && err != storage.ErrNoRows {
		return err
	}
	if len(existingPubs) > 0 {
		return nil
	}

	post, err := m.GetPost(tenantId, postId)
	if err != nil {
		return err
	}

	currentState := State(post.State)
	now := time.Now()

	if currentState == StateApproved {
		if err := m.transitionPost(tenantId, postId, "system", StatePublished, ""); err != nil {
			return err
		}
	} else if currentState != StatePublished {
		return ErrInvalidTransition
	}

	pubId := m.deps.IDs.NewID()
	pub := &Publication{
		Id:          pubId,
		TenantId:    tenantId,
		PostId:      postId,
		Channel:     channel,
		ExternalRef: externalRef,
		PublishedAt: now,
	}
	if err := pub.Validate(model.ActionCreate); err != nil {
		return err
	}
	if err := m.db.Create(pub); err != nil {
		return err
	}

	if m.deps.Publisher != nil {
		m.deps.Publisher.Publish(events.Event{
			Topic:   EventPostPublished,
			Payload: pub,
		})
	}

	return nil
}

func (m *Module) Retire(tenantId, postId, actorId string) error {
	return m.transitionPost(tenantId, postId, actorId, StateRetired, "")
}

func (m *Module) transitionPost(tenantId, postId, actorId string, targetState State, reason string) error {
	post, err := m.GetPost(tenantId, postId)
	if err != nil {
		return err
	}

	currentState := State(post.State)
	if !IsValidTransition(currentState, targetState) {
		return ErrInvalidTransition
	}

	now := time.Now()
	transId := m.deps.IDs.NewID()
	trans := &PostTransition{
		Id:        transId,
		TenantId:  tenantId,
		PostId:    postId,
		FromState: int64(currentState),
		ToState:   int64(targetState),
		ActorId:   actorId,
		Reason:    reason,
		At:        now,
	}
	if err := trans.Validate(model.ActionCreate); err != nil {
		return err
	}
	if err := m.db.Create(trans); err != nil {
		return err
	}

	post.State = int64(targetState)
	post.UpdatedAt = now
	if targetState == StatePublished && post.PublishedAt == 0 {
		post.PublishedAt = now
	}

	if err := post.Validate(model.ActionUpdate); err != nil {
		return err
	}
	if err := m.db.Update(post, orm.Eq("tenant_id", tenantId), orm.Eq("id", postId)); err != nil {
		return err
	}

	if m.deps.Publisher != nil {
		m.deps.Publisher.Publish(events.Event{
			Topic:   EventPostTransitioned,
			Payload: trans,
		})
	}

	return nil
}

func (m *Module) GetPost(tenantId, id string) (*Post, error) {
	post := &Post{}
	err := m.db.Query(post).
		Where("tenant_id").Eq(tenantId).
		Where("id").Eq(id).
		ReadOne()
	if err != nil {
		if err == orm.ErrNotFound || err == storage.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return post, nil
}

func (m *Module) UpsertPost(tenantId string, input *Post) (*Post, error) {
	if input == nil {
		return nil, fmt.Err("editorial: post cannot be nil")
	}
	input.TenantId = tenantId

	now := time.Now()

	if input.Slug != "" {
		var matches PostList
		err := m.db.Query(&Post{}).
			Where("tenant_id").Eq(tenantId).
			Where("slug").Eq(input.Slug).
			ReadAll(
				func() model.Model { return &Post{} },
				func(rec model.Model) { matches = append(matches, rec.(*Post)) },
			)
		if err == nil && len(matches) > 0 {
			for i := 0; i < len(matches); i++ {
				if matches[i].Id != input.Id {
					return nil, ErrAlreadyExists
				}
			}
		}
	}

	var existing *Post
	if input.Id != "" {
		var err error
		existing, err = m.GetPost(tenantId, input.Id)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
	}

	if existing != nil {
		input.State = existing.State
		if input.CreatedAt == 0 {
			input.CreatedAt = existing.CreatedAt
		}
		if input.PublishedAt == 0 {
			input.PublishedAt = existing.PublishedAt
		}
		input.UpdatedAt = now

		if err := input.Validate(model.ActionUpdate); err != nil {
			return nil, err
		}
		if err := m.db.Update(input, orm.Eq("tenant_id", tenantId), orm.Eq("id", input.Id)); err != nil {
			return nil, err
		}
		return input, nil
	}

	if input.Id == "" {
		input.Id = m.deps.IDs.NewID()
	}
	input.State = int64(StateDraft)
	input.CreatedAt = now
	input.UpdatedAt = now

	if err := input.Validate(model.ActionCreate); err != nil {
		return nil, err
	}
	if err := m.db.Create(input); err != nil {
		return nil, err
	}

	return input, nil
}

func (m *Module) DeletePost(tenantId, id string) error {
	_, err := m.GetPost(tenantId, id)
	if err != nil {
		return err
	}
	err = m.db.Delete(&Post{}, orm.Eq("tenant_id", tenantId), orm.Eq("id", id))
	if err != nil {
		if err == orm.ErrNotFound || err == storage.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (m *Module) ListPosts(args *ListPostsArgs) (PostList, error) {
	qb := m.db.Query(&Post{}).
		Where("tenant_id").Eq(args.TenantId)
	if args.HasStateFilter {
		qb = qb.Where("state").Eq(args.State)
	}
	if args.AuthorId != "" {
		qb = qb.Where("author_id").Eq(args.AuthorId)
	}
	if args.Limit > 0 {
		qb = qb.Limit(int(args.Limit))
	}
	if args.Offset > 0 {
		qb = qb.Offset(int(args.Offset))
	}

	var list PostList
	err := qb.ReadAll(
		func() model.Model { return &Post{} },
		func(rec model.Model) { list = append(list, rec.(*Post)) },
	)
	if err != nil && err != orm.ErrNotFound && err != storage.ErrNoRows {
		return nil, err
	}
	return list, nil
}

func (m *Module) ListTransitions(args *ListTransitionsArgs) (PostTransitionList, error) {
	qb := m.db.Query(&PostTransition{}).
		Where("tenant_id").Eq(args.TenantId).
		Where("post_id").Eq(args.PostId)

	var list PostTransitionList
	err := qb.ReadAll(
		func() model.Model { return &PostTransition{} },
		func(rec model.Model) { list = append(list, rec.(*PostTransition)) },
	)
	if err != nil && err != orm.ErrNotFound && err != storage.ErrNoRows {
		return nil, err
	}
	return list, nil
}

func (m *Module) ListPublications(args *ListPublicationsArgs) (PublicationList, error) {
	qb := m.db.Query(&Publication{}).
		Where("tenant_id").Eq(args.TenantId)
	if args.PostId != "" {
		qb = qb.Where("post_id").Eq(args.PostId)
	}
	if args.Channel != "" {
		qb = qb.Where("channel").Eq(args.Channel)
	}

	var list PublicationList
	err := qb.ReadAll(
		func() model.Model { return &Publication{} },
		func(rec model.Model) { list = append(list, rec.(*Publication)) },
	)
	if err != nil && err != orm.ErrNotFound && err != storage.ErrNoRows {
		return nil, err
	}
	return list, nil
}

package tests

import (
	"testing"

	"github.com/veltylabs/editorial"
	"github.com/tinywasm/events"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage/mem"
)

type fakeIDGen struct {
	counter int64
}

func (f *fakeIDGen) NewID() string {
	f.counter++
	return "id-" + fmt.Sprint(f.counter)
}

type fakePublisher struct {
	events []events.Event
}

func (f *fakePublisher) Publish(e events.Event) {
	f.events = append(f.events, e)
}

func setupModule(t *testing.T) (*editorial.Module, *fakePublisher) {
	t.Helper()
	db := orm.New(mem.New())
	pub := &fakePublisher{}
	deps := editorial.Deps{
		IDs:       &fakeIDGen{},
		Publisher: pub,
	}
	m, err := editorial.New(db, deps)
	if err != nil {
		t.Fatalf("editorial.New failed: %v", err)
	}
	return m, pub
}

func TestHappyPathLifecycle(t *testing.T) {
	m, pub := setupModule(t)
	tenant := "tenant-1"
	author := "author-1"
	reviewer := "reviewer-1"

	// 1. Create post
	post, err := m.UpsertPost(tenant, &editorial.Post{
		AuthorId: author,
		Slug:     "test-post",
		Title:    "Test Post",
		Body:     "Body text content",
	})
	if err != nil {
		t.Fatalf("UpsertPost failed: %v", err)
	}
	if post.State != int64(editorial.StateDraft) {
		t.Errorf("expected state Draft (%d), got %d", editorial.StateDraft, post.State)
	}

	// 2. Submit for review
	if err := m.Submit(tenant, post.Id, author); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}
	fetched, err := m.GetPost(tenant, post.Id)
	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}
	if fetched.State != int64(editorial.StateInReview) {
		t.Errorf("expected state InReview (%d), got %d", editorial.StateInReview, fetched.State)
	}

	// 3. Approve
	if err := m.Approve(tenant, post.Id, reviewer); err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	fetched, err = m.GetPost(tenant, post.Id)
	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}
	if fetched.State != int64(editorial.StateApproved) {
		t.Errorf("expected state Approved (%d), got %d", editorial.StateApproved, fetched.State)
	}

	// 4. Mark Published
	if err := m.MarkPublished(tenant, post.Id, editorial.ChannelWeb, "ext-1"); err != nil {
		t.Fatalf("MarkPublished failed: %v", err)
	}
	fetched, err = m.GetPost(tenant, post.Id)
	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}
	if fetched.State != int64(editorial.StatePublished) {
		t.Errorf("expected state Published (%d), got %d", editorial.StatePublished, fetched.State)
	}

	// Verify transitions log
	transitions, err := m.ListTransitions(&editorial.ListTransitionsArgs{
		TenantId: tenant,
		PostId:   post.Id,
	})
	if err != nil {
		t.Fatalf("ListTransitions failed: %v", err)
	}
	if len(transitions) != 3 {
		t.Fatalf("expected 3 transitions, got %d", len(transitions))
	}

	// Verify publications ledger
	pubs, err := m.ListPublications(&editorial.ListPublicationsArgs{
		TenantId: tenant,
		PostId:   post.Id,
	})
	if err != nil {
		t.Fatalf("ListPublications failed: %v", err)
	}
	if len(pubs) != 1 {
		t.Fatalf("expected 1 publication, got %d", len(pubs))
	}
	if pubs[0].Channel != editorial.ChannelWeb || pubs[0].ExternalRef != "ext-1" {
		t.Errorf("unexpected publication record: %+v", pubs[0])
	}

	// Verify events fired
	if len(pub.events) < 4 {
		t.Errorf("expected at least 4 events fired, got %d", len(pub.events))
	}
}

func TestForbiddenTransitions(t *testing.T) {
	m, _ := setupModule(t)
	tenant := "tenant-1"
	author := "author-1"
	reviewer := "reviewer-1"

	post, err := m.UpsertPost(tenant, &editorial.Post{
		AuthorId: author,
		Slug:     "forbidden-post",
		Title:    "Forbidden",
		Body:     "Body content",
	})
	if err != nil {
		t.Fatalf("UpsertPost failed: %v", err)
	}

	// Draft -> Approved is forbidden
	if err := m.Approve(tenant, post.Id, reviewer); err != editorial.ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition for Draft->Approved, got %v", err)
	}

	// Draft -> Published is forbidden
	if err := m.MarkPublished(tenant, post.Id, editorial.ChannelWeb, ""); err != editorial.ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition for Draft->Published, got %v", err)
	}

	// Submit -> InReview
	if err := m.Submit(tenant, post.Id, author); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// InReview -> Published is forbidden
	if err := m.MarkPublished(tenant, post.Id, editorial.ChannelWeb, ""); err != editorial.ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition for InReview->Published, got %v", err)
	}

	// InReview -> Approved -> Published -> Retire
	_ = m.Approve(tenant, post.Id, reviewer)
	_ = m.MarkPublished(tenant, post.Id, editorial.ChannelWeb, "")
	if err := m.Retire(tenant, post.Id, reviewer); err != nil {
		t.Fatalf("Retire failed: %v", err)
	}

	// Retired -> Submit is forbidden
	if err := m.Submit(tenant, post.Id, author); err != editorial.ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition for Retired->Submit, got %v", err)
	}
}

func TestStateDraftIsZeroValue(t *testing.T) {
	p := &editorial.Post{}
	if editorial.State(p.State) != editorial.StateDraft {
		t.Errorf("expected zero value of Post.State to be StateDraft, got %v", p.State)
	}
}

func TestUpsertCannotMutateState(t *testing.T) {
	m, _ := setupModule(t)
	tenant := "tenant-1"

	post, err := m.UpsertPost(tenant, &editorial.Post{
		AuthorId: "author",
		Slug:     "my-slug",
		Title:    "Title",
		Body:     "Body content",
	})
	if err != nil {
		t.Fatalf("UpsertPost failed: %v", err)
	}

	// Attempt to bypass review by passing StatePublished
	post.Title = "Updated Title"
	post.State = int64(editorial.StatePublished)

	updated, err := m.UpsertPost(tenant, post)
	if err != nil {
		t.Fatalf("UpsertPost update failed: %v", err)
	}

	if updated.State != int64(editorial.StateDraft) {
		t.Errorf("expected state to remain StateDraft (%d), but got %d", editorial.StateDraft, updated.State)
	}

	fetched, _ := m.GetPost(tenant, post.Id)
	if fetched.State != int64(editorial.StateDraft) {
		t.Errorf("DB state was updated to %d, expected StateDraft", fetched.State)
	}
}

func TestRequestChangesReasonRequired(t *testing.T) {
	m, _ := setupModule(t)
	tenant := "tenant-1"
	author := "author"
	reviewer := "reviewer"

	post, err := m.UpsertPost(tenant, &editorial.Post{
		AuthorId: author,
		Slug:     "review-post",
		Title:    "Title",
		Body:     "Body content",
	})
	if err != nil {
		t.Fatalf("UpsertPost failed: %v", err)
	}

	_ = m.Submit(tenant, post.Id, author)

	// Reason required
	err = m.RequestChanges(tenant, post.Id, reviewer, "")
	if err != editorial.ErrReasonRequired {
		t.Errorf("expected ErrReasonRequired, got %v", err)
	}

	// Valid reason
	err = m.RequestChanges(tenant, post.Id, reviewer, "Please fix typos in section 2")
	if err != nil {
		t.Fatalf("RequestChanges with reason failed: %v", err)
	}

	fetched, _ := m.GetPost(tenant, post.Id)
	if fetched.State != int64(editorial.StateChangesRequested) {
		t.Errorf("expected StateChangesRequested, got %d", fetched.State)
	}
}

func TestMarkPublishedIdempotent(t *testing.T) {
	m, _ := setupModule(t)
	tenant := "tenant-1"
	author := "author"
	reviewer := "reviewer"

	post, _ := m.UpsertPost(tenant, &editorial.Post{
		AuthorId: author,
		Slug:     "idempotent-post",
		Title:    "Title",
		Body:     "Body content",
	})
	_ = m.Submit(tenant, post.Id, author)
	_ = m.Approve(tenant, post.Id, reviewer)

	// First publish
	err1 := m.MarkPublished(tenant, post.Id, editorial.ChannelWeb, "ref-1")
	if err1 != nil {
		t.Fatalf("First MarkPublished failed: %v", err1)
	}

	// Second publish on same channel (idempotent)
	err2 := m.MarkPublished(tenant, post.Id, editorial.ChannelWeb, "ref-1")
	if err2 != nil {
		t.Fatalf("Second MarkPublished failed: %v", err2)
	}

	pubs, _ := m.ListPublications(&editorial.ListPublicationsArgs{
		TenantId: tenant,
		PostId:   post.Id,
	})
	if len(pubs) != 1 {
		t.Errorf("expected 1 publication record, got %d", len(pubs))
	}
}

func TestTenantIsolation(t *testing.T) {
	m, _ := setupModule(t)
	tenantA := "tenant-A"
	tenantB := "tenant-B"

	postA, err := m.UpsertPost(tenantA, &editorial.Post{
		AuthorId: "authorA",
		Slug:     "tenant-a-post",
		Title:    "Title A",
		Body:     "Body content A",
	})
	if err != nil {
		t.Fatalf("UpsertPost failed: %v", err)
	}

	// Tenant B cannot fetch Tenant A's post
	_, err = m.GetPost(tenantB, postA.Id)
	if err != editorial.ErrNotFound {
		t.Errorf("expected ErrNotFound when tenant B accesses tenant A post, got %v", err)
	}

	// Tenant B cannot submit Tenant A's post
	err = m.Submit(tenantB, postA.Id, "userB")
	if err != editorial.ErrNotFound {
		t.Errorf("expected ErrNotFound when tenant B submits tenant A post, got %v", err)
	}

	// Tenant B cannot delete Tenant A's post
	err = m.DeletePost(tenantB, postA.Id)
	if err != editorial.ErrNotFound {
		t.Errorf("expected ErrNotFound when tenant B deletes tenant A post, got %v", err)
	}

	// Tenant B listing posts gets empty list
	listB, err := m.ListPosts(&editorial.ListPostsArgs{TenantId: tenantB})
	if err != nil {
		t.Fatalf("ListPosts for tenant B failed: %v", err)
	}
	if len(listB) != 0 {
		t.Errorf("expected 0 posts for tenant B, got %d", len(listB))
	}
}

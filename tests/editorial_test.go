package tests

import (
	"testing"
	"github.com/veltylabs/editorial"
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/crud"
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
		TenantID: "tenant-1",
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

	// Create post
	post := editorial.Post{
		AuthorId: "author-1",
		Slug:     "test-post",
		Title:    "Test Post",
		Body:     "Body text content",
	}
	if err := m.CreatePost(crud.Caller{UserID: "author-1"}, editorial.PostList{&post}); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	// Ensure State is Draft (0)
	list, err := m.ReadPost(crud.Caller{UserID: "author-1"}, crud.ReadArgs{ID: post.Id}, nil)
	if err != nil || len(list) == 0 {
		t.Fatalf("ReadPost failed")
	}
	if list[0].State != int64(editorial.StateDraft) {
		t.Errorf("expected draft")
	}

	// Submit (Update state)
	postUpdate := editorial.Post{AuthorId: "author-1", Slug: "test-post", Title: "Test Post", Body: "Body text content", State: int64(editorial.StateInReview)}
	if err := m.UpdatePost(crud.Caller{UserID: "author-1"}, []string{post.Id}, []string{"state"}, postUpdate); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	list, _ = m.ReadPost(crud.Caller{UserID: "reviewer-1"}, crud.ReadArgs{ID: post.Id}, nil)
	if list[0].State != int64(editorial.StateInReview) {
		t.Errorf("expected in review")
	}

	// Approve
	trans := editorial.PostTransition{PostId: post.Id, ToState: int64(editorial.StateApproved)}
	if err := m.CreatePostTransition(crud.Caller{UserID: "reviewer-1"}, editorial.PostTransitionList{&trans}); err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	list, _ = m.ReadPost(crud.Caller{UserID: "reviewer-1"}, crud.ReadArgs{ID: post.Id}, nil)
	if list[0].State != int64(editorial.StateApproved) {
		t.Errorf("expected approved")
	}

	// Publish
	pubRec := editorial.Publication{Id: "pub-1", PostId: post.Id, Channel: "web", ExternalRef: "ext-1", PublishedAt: 123}
	if err := m.CreatePublication(crud.Caller{UserID: "system"}, editorial.PublicationList{&pubRec}); err != nil {
		t.Fatalf("CreatePublication failed: %v", err)
	}

	// Verify events
	if len(pub.events) < 3 {
		t.Errorf("expected at least 3 events")
	}
}

func TestSentinelTexts(t *testing.T) {
	if editorial.ErrNotFound.Error() != "editorial: not found" {
		t.Errorf("ErrNotFound changed: %q", editorial.ErrNotFound.Error())
	}
}

package tests

import (
	"testing"

	"github.com/veltylabs/editorial"
	"webtyp.com/crud"
	"webtyp.com/model"
)

func TestPost_UpdateOnlyListedFields(t *testing.T) {
	mod, _ := setupModule(t)
	if err := createPost("author-1", mod, "p1"); err != nil {
		t.Fatalf("create: %v", err)
	}
	err := callAs("author-1", nil, mod, "Post", model.Update,
		&crud.UpdateArgs{IDs: []string{"p1"}, Fields: []string{"excerpt"}, Record: &editorial.Post{Excerpt: "new excerpt"}}, nil)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	post, err := readPost("author-1", mod, "p1")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if post.Title != "Title" || post.Excerpt != "new excerpt" {
		t.Fatalf("only excerpt may change, got title %q excerpt %q", post.Title, post.Excerpt)
	}
}

func TestPost_AuthorCannotApprove(t *testing.T) {
	mod, _ := setupModule(t)
	if err := createPost("author-1", mod, "p1"); err != nil {
		t.Fatalf("create: %v", err)
	}
	err := callAs("author-1", nil, mod, "Post", model.Update,
		&crud.UpdateArgs{IDs: []string{"p1"}, Fields: []string{"state"}, Record: &editorial.Post{State: int64(editorial.StateApproved)}}, nil)
	if err == nil || failedField(err) != "state" {
		t.Fatalf("approving through an update must be a field error on state, got %v", err)
	}
	post, _ := readPost("author-1", mod, "p1")
	if post.State != int64(editorial.StateDraft) {
		t.Fatalf("the post must stay a draft, got state %d", post.State)
	}
}

func TestCreate_PublishesChange(t *testing.T) {
	mod, pub := setupModule(t)
	if err := callAs("author-1", pub, mod, "Post", model.Create,
		&crud.CreateArgs{Records: []model.Model{&editorial.Post{Id: "p1", Slug: "post-p1", Title: "Title", Body: "Body text of the post"}}}, nil); err != nil {
		t.Fatalf("create: %v", err)
	}
	topic := crud.Topic(editorial.ModelName, (&editorial.Post{}).ModelName(), model.Create)
	count := 0
	for _, e := range pub.events {
		if e.Topic != topic {
			continue
		}
		count++
		if _, ok := e.Payload.(*crud.Change); !ok {
			t.Fatalf("the created event must carry *crud.Change, got %T", e.Payload)
		}
	}
	if count != 1 {
		t.Fatalf("one created event on %q expected, got %d", topic, count)
	}
}

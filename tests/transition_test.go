package tests

import (
	"strings"
	"testing"

	"github.com/veltylabs/editorial"
	"webtyp.com/crud"
	"webtyp.com/model"
)

func TestTransition_ApproveFromInReview(t *testing.T) {
	mod, _ := setupModule(t)
	if err := createPost("author-1", mod, "p1"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := submitPost("author-1", mod, "p1"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	approval := &editorial.PostTransition{Id: "t1", PostId: "p1", ToState: int64(editorial.StateApproved)}
	if err := callAs("reviewer-1", nil, mod, "PostTransition", model.Create, &crud.CreateArgs{Records: []model.Model{approval}}, nil); err != nil {
		t.Fatalf("approve: %v", err)
	}
	post, err := readPost("reviewer-1", mod, "p1")
	if err != nil || post.State != int64(editorial.StateApproved) {
		t.Fatalf("the post must be approved, got state %d err %v", post.State, err)
	}
	var transitions editorial.PostTransitionList
	if err := callAs("reviewer-1", nil, mod, "PostTransition", model.Read, &crud.ReadArgs{}, &transitions); err != nil {
		t.Fatalf("read transitions: %v", err)
	}
	found := false
	for _, tr := range transitions {
		if tr.ToState == int64(editorial.StateApproved) {
			found = true
			if tr.ActorId != "reviewer-1" || tr.FromState != int64(editorial.StateInReview) {
				t.Fatalf("the approval must record the caller and the previous state, got actor %q from %d", tr.ActorId, tr.FromState)
			}
		}
	}
	if !found {
		t.Fatalf("no approval transition stored among %d", len(transitions))
	}
}

func TestTransition_ChangesRequestedNeedsReason(t *testing.T) {
	mod, _ := setupModule(t)
	if err := createPost("author-1", mod, "p1"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := submitPost("author-1", mod, "p1"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	request := &editorial.PostTransition{Id: "t1", PostId: "p1", ToState: int64(editorial.StateChangesRequested)}
	err := callAs("reviewer-1", nil, mod, "PostTransition", model.Create, &crud.CreateArgs{Records: []model.Model{request}}, nil)
	if err == nil || failedField(err) != "reason" {
		t.Fatalf("asking for changes without a reason must be a field error on reason, got %v", err)
	}
}

func TestTransition_NoUpdateOrDelete(t *testing.T) {
	mod, _ := setupModule(t)
	for _, action := range []model.Action{model.Update, model.Delete} {
		err := callAs("reviewer-1", nil, mod, "PostTransition", action, &crud.ReadArgs{ID: "t1"}, nil)
		if err == nil || !strings.Contains(err.Error(), "unknown operation") {
			t.Fatalf("action %v on a transition must not exist, got %v", action, err)
		}
	}
}

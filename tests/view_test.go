package tests

import (
	"testing"

	"github.com/veltylabs/editorial"
	"webtyp.com/model"
	"webtyp.com/router/mock"
)

func TestViewPresenter(t *testing.T) {
	caller := &mock.Caller{}
	presenter := editorial.NewView(caller)

	if presenter == nil {
		t.Fatalf("editorial.NewView returned nil")
	}

	if presenter.Title() != "Posts" {
		t.Errorf("expected Title 'Posts', got '%s'", presenter.Title())
	}

	if presenter.SearchPlaceholder() != "Search posts..." {
		t.Errorf("expected SearchPlaceholder 'Search posts...', got '%s'", presenter.SearchPlaceholder())
	}
}

func TestRowsDeclaredOnFields(t *testing.T) {
	post := &editorial.Post{
		Id:    "p1",
		Title: "My Post",
		Slug:  "my-post",
	}
	rPost := model.RowOf(post)
	if rPost.ID != "p1" || rPost.Title != "My Post" || rPost.Detail != "my-post" {
		t.Errorf("unexpected Post RowOf projection: %+v", rPost)
	}

	pub := &editorial.Publication{
		Id:          "pub1",
		Channel:     "web",
		ExternalRef: "ref1",
	}
	rPub := model.RowOf(pub)
	if rPub.ID != "pub1" || rPub.Title != "web" || rPub.Detail != "ref1" {
		t.Errorf("unexpected Publication RowOf projection: %+v", rPub)
	}

	trans := &editorial.PostTransition{
		Id:      "t1",
		ActorId: "actor1",
		Reason:  "Approved",
	}
	rTrans := model.RowOf(trans)
	if rTrans.ID != "t1" || rTrans.Title != "actor1" || rTrans.Detail != "Approved" {
		t.Errorf("unexpected PostTransition RowOf projection: %+v", rTrans)
	}
}

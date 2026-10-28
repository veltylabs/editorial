package tests

import (
	"testing"

	"github.com/veltylabs/editorial"
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

func TestItemsDeclaredOnFields(t *testing.T) {
	post := &editorial.Post{
		Id:    "p1",
		Title: "My Post",
		Slug:  "my-post",
	}
	rPost := post.Item()
	if rPost.ID != "p1" || rPost.Label != "My Post" || rPost.Description != "my-post" {
		t.Errorf("unexpected Post Item projection: %+v", rPost)
	}

	pub := &editorial.Publication{
		Id:          "pub1",
		Channel:     "web",
		ExternalRef: "ref1",
	}
	rPub := pub.Item()
	if rPub.ID != "pub1" || rPub.Label != "web" || rPub.Description != "ref1" {
		t.Errorf("unexpected Publication Item projection: %+v", rPub)
	}

	trans := &editorial.PostTransition{
		Id:      "t1",
		ActorId: "actor1",
		Reason:  "Approved",
	}
	rTrans := trans.Item()
	if rTrans.ID != "t1" || rTrans.Label != "actor1" || rTrans.Description != "Approved" {
		t.Errorf("unexpected PostTransition Item projection: %+v", rTrans)
	}
}

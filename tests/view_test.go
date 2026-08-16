package tests

import (
	"testing"

	"github.com/veltylabs/editorial"
	"github.com/tinywasm/router/mock"
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

func TestItemizers(t *testing.T) {
	post := &editorial.Post{
		Id:    "p1",
		Title: "My Post",
		Slug:  "my-post",
	}
	item := post.Item()
	if item.ID != "p1" || item.Label != "My Post" || item.Description != "my-post" {
		t.Errorf("unexpected Post Item projection: %+v", item)
	}

	pub := &editorial.Publication{
		Id:          "pub1",
		Channel:     "web",
		ExternalRef: "ref1",
	}
	pubItem := pub.Item()
	if pubItem.ID != "pub1" || pubItem.Label != "web" || pubItem.Description != "ref1" {
		t.Errorf("unexpected Publication Item projection: %+v", pubItem)
	}

	trans := &editorial.PostTransition{
		Id:      "t1",
		ActorId: "actor1",
		Reason:  "Approved",
	}
	transItem := trans.Item()
	if transItem.ID != "t1" || transItem.Label != "actor1" || transItem.Description != "Approved" {
		t.Errorf("unexpected PostTransition Item projection: %+v", transItem)
	}
}

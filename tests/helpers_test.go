package tests

import (
	"github.com/veltylabs/editorial"
	"webtyp.com/crud"
	"webtyp.com/events"
	"webtyp.com/model"
	"webtyp.com/router"
)

// callAs runs one operation of the module as the given user, through the same loopback the demo uses, and
// answers the error of the callback (nil when it succeeded).
func callAs(user string, pub events.Publisher, mod *editorial.Module, entity string, action model.Action, args model.Encodable, into model.Decodable) error {
	var callErr error
	crud.LoopbackActingAs(pub, user, mod).Call(
		router.Operation{Module: editorial.ModelName, Entity: entity, Action: action},
		args, into, func(err error) { callErr = err })
	return callErr
}

// failedField answers the field a refusal points at, or "" when the error names none.
func failedField(err error) string {
	if f, ok := err.(interface{ FailedField() string }); ok {
		return f.FailedField()
	}
	return ""
}

// createPost stores a draft post with the given id as the given user.
func createPost(user string, mod *editorial.Module, id string) error {
	post := &editorial.Post{Id: id, Slug: "post-" + id, Title: "Title", Excerpt: "old excerpt", Body: "Body text of the post"}
	return callAs(user, nil, mod, "Post", model.Create, &crud.CreateArgs{Records: []model.Model{post}}, nil)
}

// submitPost moves a draft post to in review as its author.
func submitPost(user string, mod *editorial.Module, id string) error {
	return callAs(user, nil, mod, "Post", model.Update,
		&crud.UpdateArgs{IDs: []string{id}, Fields: []string{"state"}, Record: &editorial.Post{State: int64(editorial.StateInReview)}}, nil)
}

// readPost answers the stored post.
func readPost(user string, mod *editorial.Module, id string) (editorial.Post, error) {
	var list editorial.PostList
	err := callAs(user, nil, mod, "Post", model.Read, &crud.ReadArgs{ID: id}, &list)
	if err != nil || len(list) == 0 {
		return editorial.Post{}, err
	}
	return *list[0], nil
}

package editorial

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const titlePosts = "Posts"
const searchPlaceholderPosts = "Search posts..."

// NewView builds the posts Presenter — the tech-agnostic engine a renderer
// (crudview, or any other) wraps. This module builds it (view + model + router
// only); the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(view.NewRecords().Caller(caller),
		view.Ops{Module: ModelName, List: OpListPosts, Save: OpUpsertPost, Delete: OpDeletePost},
		func() model.ModelSlice { return &PostList{} })
	return view.New(b, &Post{},
		view.WithTitle(titlePosts),
		view.WithSearchPlaceholder(searchPlaceholderPosts))
}

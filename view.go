package editorial

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const titlePosts = "Posts"

func (p *Post) Item() view.Item {
	return view.Item{
		ID:          p.Id,
		Label:       p.Title,
		Description: p.Slug,
	}
}

func (p *Publication) Item() view.Item {
	return view.Item{
		ID:          p.Id,
		Label:       p.Channel,
		Description: p.ExternalRef,
	}
}

func (t *PostTransition) Item() view.Item {
	return view.Item{
		ID:          t.Id,
		Label:       t.ActorId,
		Description: t.Reason,
	}
}

// NewView builds the posts Presenter — the tech-agnostic engine a renderer
// (crudview, or any other) wraps. This module builds it (view + model + router
// only); the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{List: OpListPosts, Save: OpUpsertPost, Delete: OpDeletePost},
		func() model.ModelSlice { return &PostList{} })
	return view.New(b, &Post{}, view.WithTitle(titlePosts))
}

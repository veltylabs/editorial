package editorial

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
	"github.com/tinywasm/view"
)

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

func NewView(caller router.Caller) view.Presenter {
	return view.New(
		caller,
		&Post{},
		OpListPosts,
		func() model.ModelSlice { return &PostList{} },
		view.WithTitle("Posts"),
		view.WithSearchPlaceholder("Search posts..."),
		view.WithSaveOp(OpUpsertPost),
		view.WithDeleteOp(OpDeletePost),
	)
}

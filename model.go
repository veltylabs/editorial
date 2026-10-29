package editorial

import (
	"webtyp.com/input"
	"webtyp.com/model"
)

const ChannelWeb = "web"

var PostModel = model.Definition{
	Name: "Post",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, DB: &model.FieldDB{}},
		{Name: "author_id", Type: model.Text(), NotNull: true, DB: &model.FieldDB{}},
		{Name: "slug", Type: input.Text(), Label: "Page address", Help: "Appears in this post's web link. Lowercase words separated by dashes, e.g. new-hours-announcement.", NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-'}}, Item: model.ItemDescription, DB: &model.FieldDB{}},
		{Name: "title", Type: input.Text(), Label: "Title", NotNull: true, Item: model.ItemLabel, DB: &model.FieldDB{}},
		{Name: "excerpt", Type: input.Textarea(), Label: "Excerpt", DB: &model.FieldDB{}},
		{Name: "body", Type: input.Textarea(), Label: "Body", NotNull: true, DB: &model.FieldDB{}},
		{Name: "cover_image", Type: input.Text(), Label: "Cover image", Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'/', '.', '-', '_'}}, DB: &model.FieldDB{}},
		{Name: "state", Type: model.Int(), DB: &model.FieldDB{}},
		{Name: "published_at", Type: model.NanoTime(), OmitEmpty: true, DB: &model.FieldDB{}},
		{Name: "created_at", Type: model.NanoTime(), DB: &model.FieldDB{}},
		{Name: "updated_at", Type: model.NanoTime(), DB: &model.FieldDB{}},
	},
}

var PostTransitionModel = model.Definition{
	Name: "PostTransition",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, DB: &model.FieldDB{}},
		{Name: "post_id", Type: model.Text(), NotNull: true, Ref: &PostModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "from_state", Type: model.Int(), DB: &model.FieldDB{}},
		{Name: "to_state", Type: model.Int(), DB: &model.FieldDB{}},
		{Name: "actor_id", Type: model.Text(), NotNull: true, Item: model.ItemLabel, DB: &model.FieldDB{}},
		{Name: "reason", Type: input.Textarea(), Label: "Reason", OmitEmpty: true, Item: model.ItemDescription, DB: &model.FieldDB{}},
		{Name: "at", Type: model.NanoTime(), NotNull: true, DB: &model.FieldDB{}},
	},
}

var PublicationModel = model.Definition{
	Name: "Publication",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, DB: &model.FieldDB{}},
		{Name: "post_id", Type: model.Text(), NotNull: true, Ref: &PostModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "channel", Type: model.Text(), NotNull: true, Item: model.ItemLabel, DB: &model.FieldDB{}},
		{Name: "external_ref", Type: model.Text(), OmitEmpty: true, Item: model.ItemDescription, DB: &model.FieldDB{}},
		{Name: "published_at", Type: model.NanoTime(), NotNull: true, DB: &model.FieldDB{}},
	},
}

var ListPostsArgsModel = model.Definition{
	Name: "ListPostsArgs",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "state", Type: model.Int(), OmitEmpty: true},
		// has_state_filter distingue "sin filtro" de "filtrar por StateDraft":
		// StateDraft es 0, el mismo zero value que "state no fue puesto", así
		// que state != 0 nunca podría pedir explícitamente los borradores.
		{Name: "has_state_filter", Type: model.Bool()},
		{Name: "author_id", Type: model.Text(), OmitEmpty: true},
		{Name: "limit", Type: model.Int()},
		{Name: "offset", Type: model.Int()},
	},
}

var GetPostArgsModel = model.Definition{
	Name: "GetPostArgs",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "id", Type: model.Text(), NotNull: true},
	},
}



var PostActionArgsModel = model.Definition{
	Name: "PostActionArgs",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "post_id", Type: model.Text(), NotNull: true, Ref: &PostModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "actor_id", Type: model.Text(), NotNull: true},
	},
}

var RequestChangesArgsModel = model.Definition{
	Name: "RequestChangesArgs",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "post_id", Type: model.Text(), NotNull: true, Ref: &PostModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "actor_id", Type: model.Text(), NotNull: true},
		{Name: "reason", Type: input.Textarea(), Label: "Reason", NotNull: true},
	},
}

var ListPublicationsArgsModel = model.Definition{
	Name: "ListPublicationsArgs",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "post_id", Type: model.Text(), OmitEmpty: true, Ref: &PostModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "channel", Type: model.Text(), OmitEmpty: true},
	},
}

var ListTransitionsArgsModel = model.Definition{
	Name: "ListTransitionsArgs",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "post_id", Type: model.Text(), NotNull: true, Ref: &PostModel, DB: &model.FieldDB{RefColumn: "id"}},
	},
}

package editorial

import (
	"webtyp.com/input"
	"webtyp.com/model"
)

const ChannelWeb = "web"

var PostModel = model.Definition{
	Allow: model.Create | model.Read | model.Update | model.Delete,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.AddMyBusinessRule,
		Update: model.AddMyBusinessRule,
		Delete: model.HaveNoBusinessRule,
	},
	Name: "Post",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true, DB: &model.FieldDB{}},
		{Name: "author_id", Type: model.Text(), NotNull: true, ServerOnly: true, DB: &model.FieldDB{}},
		{Name: "slug", Type: input.Text(), Label: "Page address", Help: "Appears in this post's web link. Lowercase words separated by dashes, e.g. new-hours-announcement.", NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-'}}, Item: model.ItemDescription, DB: &model.FieldDB{}},
		{Name: "title", Type: input.Text(), Label: "Title", NotNull: true, Item: model.ItemLabel, DB: &model.FieldDB{}},
		{Name: "excerpt", Type: input.Textarea(), Label: "Excerpt", DB: &model.FieldDB{}},
		{Name: "body", Type: input.Textarea(), Label: "Body", NotNull: true, DB: &model.FieldDB{}},
		{Name: "cover_image", Type: input.Text(), Label: "Cover image", Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'/', '.', '-', '_'}}, DB: &model.FieldDB{}},
		{Name: "state", Type: model.Int(), DB: &model.FieldDB{}},
		{Name: "published_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true, DB: &model.FieldDB{}},
		{Name: "created_at", Type: model.NanoTime(), ServerOnly: true, DB: &model.FieldDB{}},
		{Name: "updated_at", Type: model.NanoTime(), ServerOnly: true, DB: &model.FieldDB{}},
	},
}

var PostTransitionModel = model.Definition{
	Allow: model.Create | model.Read,
	Rules: model.Rules{
		Create: model.AddMyBusinessRule,
		Read:   model.HaveNoBusinessRule,
	},
	Name: "PostTransition",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true, DB: &model.FieldDB{}},
		{Name: "post_id", Type: model.Text(), NotNull: true, Ref: &PostModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "from_state", Type: model.Int(), ServerOnly: true, DB: &model.FieldDB{}},
		{Name: "to_state", Type: model.Int(), DB: &model.FieldDB{}},
		{Name: "actor_id", Type: model.Text(), NotNull: true, ServerOnly: true, Item: model.ItemLabel, DB: &model.FieldDB{}},
		{Name: "reason", Type: input.Textarea(), Label: "Reason", OmitEmpty: true, Item: model.ItemDescription, DB: &model.FieldDB{}},
		{Name: "at", Type: model.NanoTime(), NotNull: true, ServerOnly: true, DB: &model.FieldDB{}},
	},
}

var PublicationModel = model.Definition{
	Allow: model.Read,
	Rules: model.Rules{
		Read: model.HaveNoBusinessRule,
	},
	Name: "Publication",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true, DB: &model.FieldDB{}},
		{Name: "post_id", Type: model.Text(), NotNull: true, Ref: &PostModel, DB: &model.FieldDB{RefColumn: "id"}},
		{Name: "channel", Type: model.Text(), NotNull: true, Item: model.ItemLabel, DB: &model.FieldDB{}},
		{Name: "external_ref", Type: model.Text(), OmitEmpty: true, Item: model.ItemDescription, DB: &model.FieldDB{}},
		{Name: "published_at", Type: model.NanoTime(), NotNull: true, DB: &model.FieldDB{}},
	},
}









var Definition = model.ModuleDefinition{
	Name: "editorial",
	Entities: []model.Definition{
		PostModel,
		PostTransitionModel,
		PublicationModel,
	},
}

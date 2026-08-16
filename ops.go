package editorial

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
)

var _ router.OpModule = (*Module)(nil)

const (
	ResourcePost        model.Resource = "post"
	ResourcePostReview  model.Resource = "post_review"
	ResourcePublication model.Resource = "publication"
)

const (
	OpListPosts        = "list_posts"
	OpGetPost          = "get_post"
	OpUpsertPost       = "upsert_post"
	OpDeletePost       = "delete_post"
	OpSubmitPost       = "submit_post"
	OpApprovePost      = "approve_post"
	OpRequestChanges   = "request_changes"
	OpRetirePost       = "retire_post"
	OpListPublications = "list_publications"
	OpListTransitions  = "list_transitions"
)

func (m *Module) ModelName() string {
	return "editorial"
}

func mapErrorStatus(err error) int {
	if err == nil {
		return 200
	}
	if err == ErrNotFound {
		return 404
	}
	if err == ErrAlreadyExists {
		return 409
	}
	if err == ErrInvalidTransition || err == ErrReasonRequired {
		return 400
	}
	return 500
}

func (m *Module) MountOps(reg router.OpRegistry) {
	reg.Op(OpListPosts, func(ctx router.Context) {
		var args ListPostsArgs
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := args.Validate(model.ActionRead); err != nil {
			ctx.WriteStatus(400)
			return
		}
		list, err := m.ListPosts(&args)
		if err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.Encode(&list)
	}).Requires(ResourcePost, model.Read).Accepts(&ListPostsArgs{})

	reg.Op(OpGetPost, func(ctx router.Context) {
		var args GetPostArgs
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := args.Validate(model.ActionRead); err != nil {
			ctx.WriteStatus(400)
			return
		}
		post, err := m.GetPost(args.TenantId, args.Id)
		if err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.Encode(post)
	}).Requires(ResourcePost, model.Read).Accepts(&GetPostArgs{})

	reg.Op(OpUpsertPost, func(ctx router.Context) {
		var args Post
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		post, err := m.UpsertPost(args.TenantId, &args)
		if err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.Encode(post)
	}).Requires(ResourcePost, model.Create|model.Update).Accepts(&Post{})

	reg.Op(OpDeletePost, func(ctx router.Context) {
		var args DeletePostArgs
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := args.Validate(model.ActionDelete); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := m.DeletePost(args.TenantId, args.Id); err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourcePost, model.Delete).Accepts(&DeletePostArgs{})

	reg.Op(OpSubmitPost, func(ctx router.Context) {
		var args PostActionArgs
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := args.Validate(model.ActionUpdate); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := m.Submit(args.TenantId, args.PostId, args.ActorId); err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourcePost, model.Update).Accepts(&PostActionArgs{})

	reg.Op(OpApprovePost, func(ctx router.Context) {
		var args PostActionArgs
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := args.Validate(model.ActionUpdate); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := m.Approve(args.TenantId, args.PostId, args.ActorId); err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourcePostReview, model.Update).Accepts(&PostActionArgs{})

	reg.Op(OpRequestChanges, func(ctx router.Context) {
		var args RequestChangesArgs
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := args.Validate(model.ActionUpdate); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := m.RequestChanges(args.TenantId, args.PostId, args.ActorId, args.Reason); err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourcePostReview, model.Update).Accepts(&RequestChangesArgs{})

	reg.Op(OpRetirePost, func(ctx router.Context) {
		var args PostActionArgs
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := args.Validate(model.ActionUpdate); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := m.Retire(args.TenantId, args.PostId, args.ActorId); err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourcePostReview, model.Update).Accepts(&PostActionArgs{})

	reg.Op(OpListPublications, func(ctx router.Context) {
		var args ListPublicationsArgs
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := args.Validate(model.ActionRead); err != nil {
			ctx.WriteStatus(400)
			return
		}
		pubs, err := m.ListPublications(&args)
		if err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.Encode(&pubs)
	}).Requires(ResourcePublication, model.Read).Accepts(&ListPublicationsArgs{})

	reg.Op(OpListTransitions, func(ctx router.Context) {
		var args ListTransitionsArgs
		if err := ctx.Decode(&args); err != nil {
			ctx.WriteStatus(400)
			return
		}
		if err := args.Validate(model.ActionRead); err != nil {
			ctx.WriteStatus(400)
			return
		}
		trans, err := m.ListTransitions(&args)
		if err != nil {
			ctx.WriteStatus(mapErrorStatus(err))
			return
		}
		ctx.Encode(&trans)
	}).Requires(ResourcePost, model.Read).Accepts(&ListTransitionsArgs{})
}

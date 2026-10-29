package editorial

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/rpc"
	"webtyp.com/view"
)

var _ router.OperationModule = (*Module)(nil)

// ModelName is this module's identity: mcp.HarvestOps qualifies every op
// below as "editorial.<name>" on the wire, and view.go's NewView passes this
// same constant as view.Ops.Module so the client composes the identical
// qualified name.
const ModelName = "editorial"

var ResourcePostReview = model.FeatureResource("post_review")

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
	return ModelName
}

func (m *Module) MountOperations(reg router.OperationRegistry) {
	reg.Operation(OpListPosts, func(ctx router.Context) {
		var args ListPostsArgs
		if err := ctx.Decode(&args); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := args.Validate(model.ActionRead); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		list, err := m.ListPosts(&args)
		if err != nil {
			rpc.WriteError(ctx, err)
			return
		}
		ctx.Encode(&list)
	}).Requires(model.ResourceOf(&PostModel), model.Read).Accepts(&ListPostsArgs{})

	reg.Operation(OpGetPost, func(ctx router.Context) {
		var args GetPostArgs
		if err := ctx.Decode(&args); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := args.Validate(model.ActionRead); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		post, err := m.GetPost(args.TenantId, args.Id)
		if err != nil {
			rpc.WriteError(ctx, err)
			return
		}
		ctx.Encode(post)
	}).Requires(model.ResourceOf(&PostModel), model.Read).Accepts(&GetPostArgs{})

	view.MountSave(reg, OpUpsertPost, func(p *Post) error {
		out, err := m.UpsertPost(p.TenantId, p)
		if err == nil {
			*p = *out
		}
		return err
	}).Requires(model.ResourceOf(&PostModel), model.Create|model.Update)

	view.MountDelete(reg, OpDeletePost, func(ids []string) error {
		for _, id := range ids {
			if err := m.DeletePost("", id); err != nil {
				return err
			}
		}
		return nil
	}).Requires(model.ResourceOf(&PostModel), model.Delete)

	reg.Operation(OpSubmitPost, func(ctx router.Context) {
		var args PostActionArgs
		if err := ctx.Decode(&args); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := args.Validate(model.ActionUpdate); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := m.Submit(args.TenantId, args.PostId, args.ActorId); err != nil {
			rpc.WriteError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(model.ResourceOf(&PostModel), model.Update).Accepts(&PostActionArgs{})

	reg.Operation(OpApprovePost, func(ctx router.Context) {
		var args PostActionArgs
		if err := ctx.Decode(&args); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := args.Validate(model.ActionUpdate); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := m.Approve(args.TenantId, args.PostId, args.ActorId); err != nil {
			rpc.WriteError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourcePostReview, model.Update).Accepts(&PostActionArgs{})

	reg.Operation(OpRequestChanges, func(ctx router.Context) {
		var args RequestChangesArgs
		if err := ctx.Decode(&args); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := args.Validate(model.ActionUpdate); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := m.RequestChanges(args.TenantId, args.PostId, args.ActorId, args.Reason); err != nil {
			rpc.WriteError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourcePostReview, model.Update).Accepts(&RequestChangesArgs{})

	reg.Operation(OpRetirePost, func(ctx router.Context) {
		var args PostActionArgs
		if err := ctx.Decode(&args); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := args.Validate(model.ActionUpdate); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := m.Retire(args.TenantId, args.PostId, args.ActorId); err != nil {
			rpc.WriteError(ctx, err)
			return
		}
		ctx.WriteStatus(200)
	}).Requires(ResourcePostReview, model.Update).Accepts(&PostActionArgs{})

	reg.Operation(OpListPublications, func(ctx router.Context) {
		var args ListPublicationsArgs
		if err := ctx.Decode(&args); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := args.Validate(model.ActionRead); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		pubs, err := m.ListPublications(&args)
		if err != nil {
			rpc.WriteError(ctx, err)
			return
		}
		ctx.Encode(&pubs)
	}).Requires(model.ResourceOf(&PublicationModel), model.Read).Accepts(&ListPublicationsArgs{})

	reg.Operation(OpListTransitions, func(ctx router.Context) {
		var args ListTransitionsArgs
		if err := ctx.Decode(&args); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		if err := args.Validate(model.ActionRead); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		trans, err := m.ListTransitions(&args)
		if err != nil {
			rpc.WriteError(ctx, err)
			return
		}
		ctx.Encode(&trans)
	}).Requires(model.ResourceOf(&PostModel), model.Read).Accepts(&ListTransitionsArgs{})
}

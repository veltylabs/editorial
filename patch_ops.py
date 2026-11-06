import re

with open('ops.go', 'r') as f:
    content = f.read()

search = """	reg.Operation(OpListPosts, func(ctx router.Context) {
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
	}).Requires(model.ResourceOf(&PostModel), model.Read).Accepts(&ListPostsArgs{})"""

replace = """	reg.Operation(OpListPosts, func(ctx router.Context) {
		var args view.ListArgs
		if err := ctx.Decode(&args); err != nil {
			rpc.WriteError(ctx, ErrCannotRead)
			return
		}
		list, err := m.listPostsView(args.Search, args.After, view.Cap(args.Limit))
		if err != nil {
			rpc.WriteError(ctx, err)
			return
		}
		ctx.Encode(&list)
	}).Requires(model.ResourceOf(&PostModel), model.Read).Accepts(&view.ListArgs{})"""

new_content = content.replace(search, replace)
if new_content == content:
    print("Failed to replace!")
else:
    with open('ops.go', 'w') as f:
        f.write(new_content)
    print("Replaced successfully!")

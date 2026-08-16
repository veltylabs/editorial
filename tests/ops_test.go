package tests

import (
	"testing"

	"github.com/veltylabs/editorial"
	"github.com/tinywasm/json"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/router/mock"
	"github.com/tinywasm/storage/mem"
)

func TestMountOpsPermissions(t *testing.T) {
	db := orm.New(mem.New())
	deps := editorial.Deps{
		IDs: &fakeIDGen{},
	}
	m, err := editorial.New(db, deps)
	if err != nil {
		t.Fatalf("editorial.New failed: %v", err)
	}

	r := &mock.Router{}
	m.MountOps(r)

	routes := r.Routes()
	if len(routes) == 0 {
		t.Fatalf("expected mounted routes, got 0")
	}

	// Verify RBAC resource permissions on routes
	expected := []struct {
		op       string
		resource model.Resource
		action   model.Action
	}{
		{editorial.OpListPosts, editorial.ResourcePost, model.Read},
		{editorial.OpGetPost, editorial.ResourcePost, model.Read},
		{editorial.OpUpsertPost, editorial.ResourcePost, model.Create | model.Update},
		{editorial.OpDeletePost, editorial.ResourcePost, model.Delete},
		{editorial.OpSubmitPost, editorial.ResourcePost, model.Update},
		{editorial.OpApprovePost, editorial.ResourcePostReview, model.Update},
		{editorial.OpRequestChanges, editorial.ResourcePostReview, model.Update},
		{editorial.OpRetirePost, editorial.ResourcePostReview, model.Update},
		{editorial.OpListPublications, editorial.ResourcePublication, model.Read},
		{editorial.OpListTransitions, editorial.ResourcePost, model.Read},
	}

	for _, exp := range expected {
		var found bool
		for _, rt := range routes {
			if rt.Path == "/"+exp.op {
				found = true
				if rt.Resource != exp.resource {
					t.Errorf("op %s: expected resource %s, got %s", exp.op, exp.resource, rt.Resource)
				}
				if rt.Action != exp.action {
					t.Errorf("op %s: expected action %v, got %v", exp.op, exp.action, rt.Action)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected op %s to be mounted", exp.op)
		}
	}
}

func makeOpCtx(userID string, payload model.Encodable) *mock.Context {
	ctx := &mock.Context{}
	ctx.SetUserID(userID)
	if payload != nil {
		var body []byte
		_ = json.Encode(payload, &body)
		ctx.InBody = body
	}
	return ctx
}

func TestOpsConsumerLifecycle(t *testing.T) {
	db := orm.New(mem.New())
	deps := editorial.Deps{
		IDs: &fakeIDGen{},
	}
	m, err := editorial.New(db, deps)
	if err != nil {
		t.Fatalf("editorial.New failed: %v", err)
	}

	r := &mock.Router{}
	r.Configure(mock.Config{
		Authorize: func(userID string, res model.Resource, act model.Action) bool {
			return userID != ""
		},
	})
	m.MountOps(r)

	tenant := "tenant-ops"
	author := "author-ops"
	reviewer := "reviewer-ops"

	// 1. Upsert Post
	upsertArgs := &editorial.Post{
		TenantId: tenant,
		AuthorId: author,
		Slug:     "ops-post",
		Title:    "Ops Post Title",
		Body:     "Ops Post Body Content",
	}
	ctx1 := makeOpCtx(author, upsertArgs)
	r.Invoke("OP", "/"+editorial.OpUpsertPost, ctx1)
	if ctx1.Status != 200 && ctx1.Status != 0 {
		t.Fatalf("OpUpsertPost failed with status %d", ctx1.Status)
	}

	var created editorial.Post
	if err := json.Decode(ctx1.ResponseBody(), &created); err != nil {
		t.Fatalf("Decode response failed: %v", err)
	}
	if created.Id == "" {
		t.Fatalf("expected created.Id to be set")
	}

	// 2. Submit Post
	submitArgs := &editorial.PostActionArgs{
		TenantId: tenant,
		PostId:   created.Id,
		ActorId:  author,
	}
	ctx2 := makeOpCtx(author, submitArgs)
	r.Invoke("OP", "/"+editorial.OpSubmitPost, ctx2)
	if ctx2.Status != 200 {
		t.Fatalf("OpSubmitPost status %d", ctx2.Status)
	}

	// 3. Approve Post
	approveArgs := &editorial.PostActionArgs{
		TenantId: tenant,
		PostId:   created.Id,
		ActorId:  reviewer,
	}
	ctx3 := makeOpCtx(reviewer, approveArgs)
	r.Invoke("OP", "/"+editorial.OpApprovePost, ctx3)
	if ctx3.Status != 200 {
		t.Fatalf("OpApprovePost status %d", ctx3.Status)
	}

	// 4. Get Post via Op
	getArgs := &editorial.GetPostArgs{
		TenantId: tenant,
		Id:       created.Id,
	}
	ctx4 := makeOpCtx(author, getArgs)
	r.Invoke("OP", "/"+editorial.OpGetPost, ctx4)
	if ctx4.Status != 200 && ctx4.Status != 0 {
		t.Fatalf("OpGetPost status %d", ctx4.Status)
	}

	var fetched editorial.Post
	if err := json.Decode(ctx4.ResponseBody(), &fetched); err != nil {
		t.Fatalf("Decode fetched post failed: %v", err)
	}
	if fetched.State != int64(editorial.StateApproved) {
		t.Errorf("expected state Approved, got %d", fetched.State)
	}
}

package tests

import (
	"testing"
	stdjson "encoding/json"


	"github.com/veltylabs/editorial"
	"webtyp.com/json"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router/mock"
	"webtyp.com/storage/mem"
)

func TestMountOperationsPermissions(t *testing.T) {
	db := orm.New(mem.New())
	deps := editorial.Deps{
		IDs: &fakeIDGen{},
	}
	m, err := editorial.New(db, deps)
	if err != nil {
		t.Fatalf("editorial.New failed: %v", err)
	}

	r := &mock.Router{}
	m.MountOperations(r)

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
		{editorial.OpListPosts, model.ResourceOf(&editorial.PostModel), model.Read},
		{editorial.OpGetPost, model.ResourceOf(&editorial.PostModel), model.Read},
		{editorial.OpUpsertPost, model.ResourceOf(&editorial.PostModel), model.Create | model.Update},
		{editorial.OpDeletePost, model.ResourceOf(&editorial.PostModel), model.Delete},
		{editorial.OpSubmitPost, model.ResourceOf(&editorial.PostModel), model.Update},
		{editorial.OpApprovePost, editorial.ResourcePostReview, model.Update},
		{editorial.OpRequestChanges, editorial.ResourcePostReview, model.Update},
		{editorial.OpRetirePost, editorial.ResourcePostReview, model.Update},
		{editorial.OpListPublications, model.ResourceOf(&editorial.PublicationModel), model.Read},
		{editorial.OpListTransitions, model.ResourceOf(&editorial.PostModel), model.Read},
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
	m.MountOperations(r)

	tenant := "tenant-ops"
	author := "author-ops"
	reviewer := "reviewer-ops"

	// 1. Upsert Post
	ctx1 := makeOpCtx(author, nil)
	ctx1.InBody = []byte("{\"records\":[{\"tenant_id\":\"" + tenant + "\",\"author_id\":\"" + author + "\",\"slug\":\"ops-post\",\"title\":\"Ops Post Title\",\"body\":\"Ops Post Body Content\"}]}")
	r.Invoke("OP", "/"+editorial.OpUpsertPost, ctx1)
	if ctx1.Status != 200 && ctx1.Status != 0 {
		t.Fatalf("OpUpsertPost failed with status %d %s", ctx1.Status, string(ctx1.ResponseBody()))
	}
	// We must decode array of maps because it's a slice. Or via MountSave it encodes the returned record.

	var responseData map[string]interface{}
	if err := stdjson.Unmarshal(ctx1.ResponseBody(), &responseData); err != nil {
		t.Fatalf("Decode response failed: %v", err)
	}
	createdId, ok := responseData["id"].(string)
	if !ok || createdId == "" {
		t.Fatalf("expected createdId")
	}

	// 2. Submit Post
	submitArgs := &editorial.PostActionArgs{
		TenantId: tenant,
		PostId:   createdId,
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
		PostId:   createdId,
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
		Id:       createdId,
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

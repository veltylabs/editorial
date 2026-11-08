package tests

import (
	"testing"

	"github.com/veltylabs/editorial"
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router/mock"
	"webtyp.com/storage/mem"
	"webtyp.com/fmt"
)

type fakeIDGen2 struct {
	counter int64
}

func (f *fakeIDGen2) NewID() string {
	f.counter++
	return "id-" + fmt.Sprint(f.counter)
}

func TestMountOperationsPermissions(t *testing.T) {
	db := orm.New(mem.New())
	deps := editorial.Deps{
		TenantID: "tenant-ops",
		IDs: &fakeIDGen2{},
	}
	m, err := editorial.New(db, deps)
	if err != nil {
		t.Fatalf("editorial.New failed: %v", err)
	}

	r := &mock.Router{}
	if err := crud.Mount(r, nil, m); err != nil {
		t.Fatalf("crud.Mount failed: %v", err)
	}

	routes := r.Routes()
	if len(routes) == 0 {
		t.Fatalf("expected mounted routes, got 0")
	}


	expected := []struct {
		op       string
		resource model.Resource
		action   model.Action
	}{
		{"/Post_r", model.ResourceOf(&editorial.PostModel), model.Read},
		{"/Post_c", model.ResourceOf(&editorial.PostModel), model.Create},
		{"/Post_u", model.ResourceOf(&editorial.PostModel), model.Update},
		{"/Post_d", model.ResourceOf(&editorial.PostModel), model.Delete},
		{"/PostTransition_r", model.ResourceOf(&editorial.PostTransitionModel), model.Read},
		{"/PostTransition_c", model.ResourceOf(&editorial.PostTransitionModel), model.Create},
		{"/PostTransition_u", model.ResourceOf(&editorial.PostTransitionModel), model.Update},
		{"/PostTransition_d", model.ResourceOf(&editorial.PostTransitionModel), model.Delete},
		{"/Publication_r", model.ResourceOf(&editorial.PublicationModel), model.Read},
		{"/Publication_c", model.ResourceOf(&editorial.PublicationModel), model.Create},
		{"/Publication_u", model.ResourceOf(&editorial.PublicationModel), model.Update},
		{"/Publication_d", model.ResourceOf(&editorial.PublicationModel), model.Delete},
	}

	for _, exp := range expected {
		found := false
		for _, route := range routes {
			if route.Path == exp.op {
				if route.Resource != exp.resource {
					t.Errorf("route %s: expected resource %q, got %q", exp.op, exp.resource, route.Resource)
				}
				if route.Action != exp.action {
					t.Errorf("route %s: expected action %d, got %d", exp.op, exp.action, route.Action)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("route %s not mounted", exp.op)
		}
	}
}

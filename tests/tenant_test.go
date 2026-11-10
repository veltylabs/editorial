package tests

import (
	"testing"

	"github.com/veltylabs/editorial"
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

func TestTenant_ReadOtherTenantIsNotFound(t *testing.T) {
	db := orm.New(mem.New())
	modA, err := editorial.New(db, editorial.Deps{TenantID: "tenant-a", IDs: &fakeIDGen{}})
	if err != nil {
		t.Fatal(err)
	}
	modB, err := editorial.New(db, editorial.Deps{TenantID: "tenant-b", IDs: &fakeIDGen{}})
	if err != nil {
		t.Fatal(err)
	}
	if err := createPost("author-1", modA, "p1"); err != nil {
		t.Fatalf("create in tenant a: %v", err)
	}
	if _, err := readPost("author-1", modA, "p1"); err != nil {
		t.Fatalf("tenant a must read its own post: %v", err)
	}
	_, err = readPost("author-1", modB, "p1")
	if err == nil || model.KindOf(err) != model.ErrorNotFound {
		t.Fatalf("tenant b reading a post of tenant a must be ErrorNotFound, got %v", err)
	}
}

func TestMount_MatchesDefinition(t *testing.T) {
	mod, _ := setupModule(t)
	if err := crud.Mount(nil, nil, nil, mod); err != nil {
		t.Fatalf("what MountOperations mounts must match Definition(), got %v", err)
	}
}

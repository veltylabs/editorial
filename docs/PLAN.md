---
PLAN: "refactor: module review — errors via rpc.WriteError, resources from tables, screens via view.MountSave, foreign keys, translations"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 15692168558440316430
---

> Read AGENTS.md at the repo root before any change.

# PLAN — `editorial` passes the module review

## Why

Before a module enters an app it passes the 6-point review of skill **app-modules** («Before a
module enters an app»), and it adopts the pieces the framework shipped on 2026-10-29 (decisions
D22, D26, D27 of `veltylabs/mjosefa-cms/docs/MVP_DEPLOYMENT_ROADMAP.md`). Measured in this repo:

| Point | State today |
|---|---|
| Errors | domainError + writeError, 19 raw `ctx.WriteStatus(4xx/5xx)`, 1 `==` |
| Permissions | `ResourcePost`, `ResourcePostReview`, `ResourcePublication` consts |
| Save/delete of screens | the screen sends `{"records":[…]}`/`{"ids":[…]}`; these ops must read exactly that: `OpUpsertPost`, `OpDeletePost` |
| Records `Item()` | present |
| Foreign keys | 6 `Ref` — check every `*_id` column |
| Translations | es filled, 7 empty (175 values) |
| Demo | none: add `web/browser/main.go` |

## APIs used (exact, from the published source)

| Symbol | Exact | Tag |
|---|---|---|
| `view.MountSave` | `func MountSave[T any, PT interface{ *T; model.Model }](reg router.OperationRegistry, op string, save func(PT) error) router.Route` | view v0.6.55 |
| `view.MountDelete` | `func MountDelete(reg router.OperationRegistry, op string, remove func(ids []string) error) router.Route` | view v0.6.55 |
| `rpc.WriteError` | `func WriteError(ctx router.Context, err error)` — status from the error kind; a 500 never sends `err.Error()` | rpc v0.2.12 |
| `model.ResourceOf` | `func ResourceOf(def *Definition) Resource` — the permission of a table (`""` if the definition has no PK) | model v0.2.12 |
| `model.FeatureResource` | `func FeatureResource(name string) Resource` — a permission that is not a table; declare it once as a package `var` | model v0.2.12 |
| `model.NotFound`, `model.Conflict`, `model.Invalid` | `type X string` with `Error()` and `Kind()`; `const ErrX model.NotFound = "…"` | model v0.2.12 |
| `model.FieldError` | `struct { Field string; Err error }` — a validation failure of one field | model v0.2.12 |
| `model.KindOf` | `func KindOf(err error) ErrorKind` (`ErrorInternal`, `ErrorInvalid`, `ErrorNotFound`, `ErrorConflict`) | model v0.2.12 |

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go install webtyp.com/ormc/cmd/ormc@latest
go get webtyp.com/view@v0.6.55 webtyp.com/rpc@v0.2.12 webtyp.com/model@v0.2.12 webtyp.com/router@v0.4.16
```

## Stage 1 — errors

Delete the module's own `domainError`, `ValidationError`, `statusOf`/`mapErrorStatus` and
`writeError`. Re-type every `Err*` constant, **keeping its text and today's status**: 404 →
`model.NotFound`, 409 → `model.Conflict`, 400 → `model.Invalid`. Add
`ErrCannotRead model.Invalid = "editorial: cannot read the request"` for decode failures. A validation
wrapper around a model `Validate` error returns that error as is (it is already a
`*model.FieldError`); a hand-made field check becomes `&model.FieldError{Field: "<column>", Err: …}`.
Every handler answers errors with `rpc.WriteError(ctx, err)` — no `ctx.WriteStatus(4xx/5xx)` and no
`ctx.Write(err.Error())` left. Replace every `==`/`!=` between errors with `model.KindOf(err) ==
model.Error…` (it pulls reflection into the browser binary otherwise).

## Stage 2 — permissions

Delete every `Resource*` constant that names a table and every `Requires("…")`/`Permission: "…"`
literal: a table's permission is `model.ResourceOf(&XModel)` (same name as today when the constant
already equalled the table name). A permission that is not a table is declared once:
`var ResourceX = model.FeatureResource("x")`. Say in the PR any name that changes.

## Stage 3 — save and delete of the screens

1. In `editorial.go`, update `GetPost` to only filter `tenant_id` if non-empty (so `DeletePost("", id)` can find the post by ID):
```go
func (m *Module) GetPost(tenantId, id string) (*Post, error) {
    if id == "" {
        return nil, fmt.Err("editorial: GetPost requires an id")
    }
    post := &Post{}
    q := m.db.Query(post)
    if tenantId != "" {
        q = q.Where("tenant_id").Eq(tenantId)
    }
    err := q.Where("id").Eq(id).ReadOne()
    if err != nil {
        if orm.IsNotFound(err) || storage.IsNoRows(err) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return post, nil
}
```
2. Mount save and delete:
- `OpUpsertPost`: `view.MountSave(reg, OpUpsertPost, func(p *Post) error { out, err := m.UpsertPost(p.TenantId, p); if err == nil { *p = *out }; return err }).Requires(model.ResourceOf(&PostModel), model.Create|model.Update)`.
- `OpDeletePost`: `view.MountDelete(reg, OpDeletePost, func(ids []string) error { for _, id := range ids { if err := m.DeletePost("", id); err != nil { return err } }; return nil }).Requires(model.ResourceOf(&PostModel), model.Delete)`.
3. Delete `DeletePostArgsModel` (regenerate `model_orm.go`).

## Stage 4 — records and foreign keys

Declare `Ref: &XModel, DB: &model.FieldDB{RefColumn: "id"}` on every `*_id` column that points to a
table of this module (keep `PK: true` where the column is part of the key); regenerate `model_orm.go` with the repo's
`ormc`. A column pointing to another module's table gets `Ref` only if `go.mod` already requires
that module.

## Stage 5 — translations

English keys only. Fill `lang.json` with **real** translations for every
language the file declares — never empty, never a placeholder. `langc sync .` → 0 pending, 0 reuse
warnings, 0 identical values.

## Stage 6 — tests and demo

- Update the tests that used deleted types or handlers.
- New `tests/screen_save_test.go`:
```go
recs := view.NewRecords()
caller := recs.Caller(loopback.ActingAs("tenant-1", "user-1", m))
pres := editorial.NewView(caller)

var saveErr error
pres.(view.Saver).Save([]model.Model{&editorial.Post{
    TenantId: "tenant-1",
    Slug:     "test-article",
    Title:    "Test Title",
    Body:     "Test Body",
    AuthorId: "user-1",
}}, func(err error) { saveErr = err })
if saveErr != nil {
    t.Fatalf("save failed: %v", saveErr)
}

items := pres.Items()
if len(items) == 0 {
    t.Fatalf("expected saved post in presenter items")
}
id := items[0].ID

var delErr error
pres.(view.Deleter).Delete([]string{id}, func(err error) { delErr = err })
if delErr != nil {
    t.Fatalf("delete failed: %v", delErr)
}
```
- Demo: `web/browser/main.go` using `storage/mem`, `loopback.ActingAs("demo", "user-1", m)`, and `editorialui.Screens(caller)`. Verify with `go build ./web/browser`. Do NOT attempt to install `webtyp` or take screenshots in headless CI; screenshots are reviewed locally upon pull.

## Stage 7 — docs (Spanish)

`README.md`: errores por `rpc.WriteError`; permisos por tabla (`model.ResourceOf`); guardar/borrar con `view.MountSave`/`MountDelete`; claves foráneas declaradas.

## Delivery check

```
GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- When the check passes, commit, push and open the PR immediately. Do not pause to ask questions.

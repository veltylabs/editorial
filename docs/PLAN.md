---
PLAN: "refactor!: migrate github.com/tinywasm -> webtyp.com + adopt view.NewCallerLister"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `editorial`: WebTyp rename + new `view.New` API

The framework moved `github.com/tinywasm/*` → `webtyp.com/*` (vanity path, every
module published). `origin/main` of this repo is still entirely on
`github.com/tinywasm/*`. Two jobs:

- **A.** The mechanical rename `github.com/tinywasm/*` → `webtyp.com/*`.
- **B.** Adopt the new `webtyp.com/view` `view.New` signature.

Module import path **stays** `github.com/veltylabs/editorial`.

This is the identical migration three sibling modules just completed —
`github.com/veltylabs/{item_catalog,device_manager,clinical_encounter}` on
`main` are working references.

---

## A. Rename `github.com/tinywasm` → `webtyp.com`

### A1. Go source

Replace import-path prefix **`github.com/tinywasm/`** → **`webtyp.com/`** in
every `.go` file (`grep -rln 'github.com/tinywasm' --include='*.go' .`). Package
selectors do not change — only the path in the `import` block.

### A2. `go.mod`

`origin/main` require block:

```
github.com/tinywasm/ddl v0.0.9
github.com/tinywasm/events v0.0.2
github.com/tinywasm/fmt v0.25.5
github.com/tinywasm/input v0.0.3
github.com/tinywasm/model v0.1.4
github.com/tinywasm/orm v0.11.6
github.com/tinywasm/router v0.1.21
github.com/tinywasm/storage v0.0.3
github.com/tinywasm/time v0.5.2
github.com/tinywasm/view v0.1.17
github.com/tinywasm/json v0.5.19
```

For **each**: `go mod edit -droprequire=github.com/tinywasm/<X>` then
`go get webtyp.com/<X>@latest`; then `go mod tidy`. `@latest` is authoritative;
current tags for reference: `ddl v0.0.15`, `events v0.0.3`, `fmt v1.0.0`,
`input v0.0.6`, `model v0.1.8`, `orm v0.12.1`, `router v0.1.31`,
`storage v0.0.7`, `time v0.5.5`, `view v0.5.2`, `json v0.5.25`. No
`github.com/tinywasm/*` left in `go.mod`; no `replace … => ../…` outside this
module.

### A3. Docs / config text

`*.md` / `*.yml` / `*.yaml`: `github.com/tinywasm/` → `github.com/webtyp/`.
Prose `TinyWasm`/`TinyWASM` → `WebTyp`. Leave `LICENSE` and upstream "TinyGo".

---

## B. New `view.New` API — use `view.NewCallerLister`

`webtyp.com/view`'s `view.New` no longer takes a `router.Caller`, an op-name
string, a slice factory, or `view.WithSaveOp` / `view.WithDeleteOp` (both
**removed**):

```go
func New(l Lister, record model.Model, opts ...Option) Presenter
type Lister interface{ List() ([]model.Model, error) }
```

Framework ships the behaviour-preserving adapter:

```go
func NewCallerLister(c router.Caller, ops Ops, newList func() model.ModelSlice) Lister
type Ops struct{ List, Save, Update, Delete string }  // Ops.List required;
// returned Lister carries exactly the write caps whose op name is non-empty
```

### Reference

- `webtyp.com/auth` → `auth/view.go` (canonical).
- `github.com/veltylabs/item_catalog` `view.go` on `main` — this exact change.

### Do — rewrite `view.go`

`NewView` keeps its **exact current signature**
(`func(caller router.Caller) view.Presenter`). Imports `webtyp.com/model`,
`webtyp.com/router`, `webtyp.com/view` stay. `origin/main` body is:

```go
return view.New(
	caller, record, OpListPosts,
	func() model.ModelSlice { return &PostList{} },
	view.WithTitle("Posts"),
	view.WithSaveOp(OpUpsertPost),
	view.WithDeleteOp(OpDeletePost),
)
```

becomes:

```go
// NewView builds the posts Presenter — the tech-agnostic engine a renderer
// (crudview, or any other) wraps. This module builds it (view + model + router
// only); the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{List: OpListPosts, Save: OpUpsertPost, Delete: OpDeletePost},
		func() model.ModelSlice { return &PostList{} })
	return view.New(b, &Post{}, view.WithTitle(titlePosts))
}
```

`titlePosts` is a new unexported constant (`const titlePosts = "Posts"`) — do
not inline. `OpListPosts`, `OpUpsertPost`, `OpDeletePost` already exist in
`ops.go` — reuse. Confirm the record type name (`&Post{}` — check `model.go`)
and the `view.Itemizer` `Item()` method stays exactly as it is.

### Tests

`NewView`'s signature is unchanged, so tests calling `NewView(fakeCaller)` keep
working. Adapt any test referencing a removed symbol (`view.WithSaveOp` /
`view.WithDeleteOp` / old `view.New` arity) minimally, preserving the assertion
intent (lists rows, is Saver, is Deleter).

---

## Verify

```
grep -rn 'github.com/tinywasm' --include='*.go' --include='go.mod' .   # empty
grep -rn 'view.WithSaveOp\|view.WithDeleteOp' .                        # empty
grep -rn '=> \.\./' go.mod                                            # empty
```

## Acceptance

- `go build ./...` → clean.
- `gotest ./...` → all green (vet, race, tests).
- No `github.com/tinywasm/*` anywhere in `*.go` / `go.mod`.
- `view.go` no longer references `view.WithSaveOp` / `view.WithDeleteOp` / the
  5-arg `view.New`.
- `NewView` still has signature `func(caller router.Caller) view.Presenter`.

## Constraints

- **No behaviour change** — path rename + adapter swap only. The `view.Ops` op
  names are the same strings the old options used.
- **No hardcoded strings** — the view title is a named constant; op names
  already are (`ops.go`).
- Keep every `//go:build` tag exactly as-is (backend + WASM shared).
- Do not populate `view.Ops.Update` — the old view had no update op.

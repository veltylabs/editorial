---
PLAN: "refactor: editorial rules answer with crud.RuleAnswer, Search stops searching text"
EXECUTOR: antigravity
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan — `github.com/veltylabs/editorial`: bring the module to the current `crud`

Orchestrated from the owner's `EXECUTION_ORDER_MASTER_PLAN` (block E). **Do not start before the versions table below has no `{{` left** (it is filled when the framework release of «RECORD_SEARCH phase 2» is published).

## 0. Context (verified in a scratch copy on 2026-11-16: every command below was run and `gotest` ended green on every destination — do not re-diagnose)

`editorial` still pins an old `webtyp.com/crud` (v0.0.11/v0.0.12). Three framework changes reached it:

1. **`crud.RuleAnswer` is a struct, and `ApplyAndPublish` is a function.** A rule that rejects returns `crud.RuleAnswer{}` (the zero value rejects); a rule that accepts returns `crud.ApplyAndPublish()`. `return 0, err` and `return crud.ApplyAndPublish, nil` no longer compile.
2. **A `Search` rule no longer receives the typed text.** `crud` itself filters every typed word on the hidden column `text_search` (accent- and case-insensitive); the rule only adds AND conditions (state, site…). The parameter `text string` is removed from the signature, and any branch that searched the text with `LIKE` is deleted.
3. **`ddl.Sync` now keeps `text_search`.** It adds the column and rebuilds it by reading rows, so the connection given to `Migrate` must be a `storage.Conn`. The migration test of this repo uses a connection that only swallows DDL and now fails (`ddl: text_search of this table must be rebuilt, and the connection cannot read rows`). Stage 5 fixes the test; `migrate.go` does not change.

Anti-footguns: do **not** touch `definition_gen.go` by hand (a later stage regenerates it); do **not** add a `replace` to `go.mod`; do **not** write `go.sum` lines (only `go get` and `go mod tidy` write them); do **not** edit with `sed`/regex/scripts — one file at a time with the editor, `gotest` at the end of each stage; do **not** change any rule's logic, only what this plan lists.

## 1. Tools and versions

Versions (filled at dispatch from the published release; a `{{` left means: stop and say so in the PR):

| Module | Tag |
|---|---|
| `webtyp.com/model` | `v0.2.24` |
| `webtyp.com/input` | `v0.2.5` |
| `webtyp.com/storage` | `v0.2.8` |
| `webtyp.com/orm` | `v0.12.42` |
| `webtyp.com/ddl` | `v0.0.57` |
| `webtyp.com/crud` | `v0.0.17` |
| `webtyp.com/definitioncheck` | `v0.0.10` |
| `webtyp.com/definitionc` | `v0.1.49` |
| `webtyp.com/view` | `v0.6.86` |

```bash
export GOPRIVATE="webtyp.com/*,github.com/veltylabs/*" GOFLAGS=-mod=mod
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/definitionc/cmd/definitionc@v0.1.49
export PATH="$PATH:$(go env GOPATH)/bin"
go get webtyp.com/model@v0.2.24 webtyp.com/input@v0.2.5 webtyp.com/storage@v0.2.8 webtyp.com/orm@v0.12.42 webtyp.com/ddl@v0.0.57 webtyp.com/crud@v0.0.17 webtyp.com/definitioncheck@v0.0.10 webtyp.com/view@v0.6.86
go mod tidy
```

Run `go vet ./...` once: the errors it prints are exactly the sites of stages 3–5 (and `definition_gen.go`, which stage 6 regenerates). `go.mod` must end with no `replace` line.

## 2. APIs used (read from the published tags above; if a signature differs, stop and name the symbol in the PR — never adapt it by guessing)

| Symbol | Exact signature | File |
|---|---|---|
| `crud.RuleAnswer` | `type RuleAnswer struct{…}`; its zero value is a rejection | `crud/result.go` |
| `crud.ApplyAndPublish` | `func ApplyAndPublish(calls ...ExternalCall) RuleAnswer` | `crud/result.go` |
| `crud.ApplyWithoutPublishing` | `func ApplyWithoutPublishing(calls ...ExternalCall) RuleAnswer` | `crud/result.go` |
| `Search` hook the generator wires | `func(a crud.CurrentUser, filter model.Model) ([]storage.Condition, error)`; your method is `func (p *X) Search(a crud.CurrentUser, filter *X, m *Module) ([]storage.Condition, error)` | `crud/contract.go` |
| `model.TextSearchFields` | `func TextSearchFields(schema []Field) []int` (indexes of the fields that enter `text_search`; empty → the table has no such column) | `model/text_search.go` |
| `model.FieldTextSearch` | `const FieldTextSearch = "text_search"` | `model/text_search.go` |
| `mem.New` | `func New() storage.Conn` (package `webtyp.com/storage/mem`) | `storage/mem/mem.go` |
| `storage.Affected` | `func Affected(n int64) Result` | `storage/` |

## 3. Stage — the rules answer with `crud.RuleAnswer` (one file at a time; `gotest` after the last)

In each file below, inside the rule's body only: every `return 0, <err>` becomes `return crud.RuleAnswer{}, <err>`, and the final `return crud.ApplyAndPublish, nil` becomes `return crud.ApplyAndPublish(), nil` (or `crud.ApplyWithoutPublishing` → `crud.ApplyWithoutPublishing()` where the file has it). Nothing else changes. Worked example (`rule_Post_update.go` of `editorial`):

```go
// before
return 0, &model.FieldError{Field: "state", Err: ErrInvalidTransition}
return crud.ApplyAndPublish, nil
// after
return crud.RuleAnswer{}, &model.FieldError{Field: "state", Err: ErrInvalidTransition}
return crud.ApplyAndPublish(), nil
```

| File | `return 0,` to change | `crud.Apply…` to call |
|---|---|---|
| `rule_PostTransition_create.go` | 5 | 1 |
| `rule_Post_create.go` | 0 | 1 |
| `rule_Post_update.go` | 3 | 1 |

Total: 8 returns and 3 calls. Check: `grep -n "return 0, " rule_*.go` → empty; `grep -nE "crud\.Apply(AndPublish|WithoutPublishing)," rule_*.go` → empty.

## 4. Stage — `Search` rules lose `text`

| File | New signature (line 1 of the function) | Also delete |
|---|---|---|
| `rule_Post_search.go` | `func (p *Post) Search(a crud.CurrentUser, filter *Post, m *Module) ([]storage.Condition, error) {` | the `if text != ""` block (3 lines) and its `storage.Like`/`orm.Like` call |

Never replace a deleted text branch with another text search: `crud` already searches `text_search`. Example (`chat_room`, `rule_chat_room_search.go`):

```go
// before
func (p *Room) Search(a crud.CurrentUser, filter *Room, text string, m *Module) ([]storage.Condition, error) {
	var conditions []storage.Condition
	if filter != nil && filter.Kind != "" {
		conditions = append(conditions, storage.Eq(Room_.Kind, filter.Kind))
	}
	if text != "" {
		conditions = append(conditions, storage.Like(Room_.Name, "%"+text+"%"))
	}
	return conditions, nil
}
// after
func (p *Room) Search(a crud.CurrentUser, filter *Room, m *Module) ([]storage.Condition, error) {
	var conditions []storage.Condition
	if filter != nil && filter.Kind != "" {
		conditions = append(conditions, storage.Eq(Room_.Kind, filter.Kind))
	}
	return conditions, nil
}
```

Check: `grep -n "text string" rule_*search.go` → empty.

## 5. Stage — the migration test reads through a `storage.Conn`

Apply this diff to `tests/migrate_test.go` (verified: with it `gotest` is green). It makes the fake connection embed `mem.New()` (so `ddl` can read rows), keeps swallowing DDL, hides `ddl`'s own bookkeeping table `text_search_state` from the recorded statements, and counts the `text_search` column among the columns of a table that has searchable fields.

```diff
@@ -4,6 +4,7 @@
 	"errors"
 	"testing"
 	"webtyp.com/storage"
+	"webtyp.com/storage/mem"
 
 	"webtyp.com/ddl"
 	"webtyp.com/model"
@@ -14,6 +15,7 @@
 // migrateConn is the database the migration sees: Exec does nothing, and
 // TableColumns answers which columns each table already has.
 type migrateConn struct {
+	storage.Conn // mem: ddl reconstruye text_search a través de ella
 	existing  []tableColumns
 	failTable string
 }
@@ -25,7 +27,12 @@
 
 var errIntrospect = errors.New("introspection failed")
 
-func (c *migrateConn) Exec(string, ...any) (storage.Result, error) { return nil, nil }
+func (c *migrateConn) Exec(query string, args ...any) (storage.Result, error) {
+	if query == "" { // en lo que recordingCompiler compila cada sentencia DDL
+		return storage.Affected(0), nil
+	}
+	return c.Conn.Exec(query, args...)
+}
 
 func (c *migrateConn) TableColumns(table string) ([]string, error) {
 	if table == c.failTable {
@@ -43,6 +50,9 @@
 type recordingCompiler struct{ stmts []ddl.Stmt }
 
 func (r *recordingCompiler) CompileDDL(s ddl.Stmt, _ model.Model) (string, []any, error) {
+	if s.Table == "text_search_state" { // tabla propia de ddl, no de este repo
+		return "", nil, nil
+	}
 	r.stmts = append(r.stmts, s)
 	return "", nil, nil
 }
@@ -57,6 +67,9 @@
 	for _, f := range m.Schema() {
 		out = append(out, f.Name)
 	}
+	if len(model.TextSearchFields(m.Schema())) > 0 {
+		out = append(out, model.FieldTextSearch)
+	}
 	return out
 }
 
@@ -71,7 +84,7 @@
 // Branch "listing works, nothing missing": only the idempotent CREATE TABLE per table, in order.
 func TestMigrate_UpToDate_OnlyEnsuresTables(t *testing.T) {
 	comp := &recordingCompiler{}
-	if err := editorial.Migrate(&migrateConn{existing: upToDate()}, comp); err != nil {
+	if err := editorial.Migrate(&migrateConn{Conn: mem.New(), existing: upToDate()}, comp); err != nil {
 		t.Fatalf("Migrate: %v", err)
 	}
 	models := migrationModels()
@@ -94,7 +107,7 @@
 	existing[len(existing)-1].columns = cols[:len(cols)-1]
 
 	comp := &recordingCompiler{}
-	if err := editorial.Migrate(&migrateConn{existing: existing}, comp); err != nil {
+	if err := editorial.Migrate(&migrateConn{Conn: mem.New(), existing: existing}, comp); err != nil {
 		t.Fatalf("Migrate: %v", err)
 	}
 	var adds []ddl.Stmt
@@ -115,7 +128,7 @@
 func TestMigrate_IntrospectionFails_ReturnsErrorAndStops(t *testing.T) {
 	first := migrationModels()[0]
 	comp := &recordingCompiler{}
-	err := editorial.Migrate(&migrateConn{existing: upToDate(), failTable: first.ModelName()}, comp)
+	err := editorial.Migrate(&migrateConn{Conn: mem.New(), existing: upToDate(), failTable: first.ModelName()}, comp)
 	if !errors.Is(err, errIntrospect) {
 		t.Fatalf("got %v, want the introspection error", err)
 	}
```

The three tests of the file keep their names and meaning. Do not touch `migrate.go`.

## 6. Stage — regenerate `definition_gen.go`

```bash
definitionc
go get webtyp.com/model@v0.2.24 webtyp.com/input@v0.2.5 webtyp.com/storage@v0.2.8 webtyp.com/orm@v0.12.42 webtyp.com/ddl@v0.0.57 webtyp.com/crud@v0.0.17 webtyp.com/definitioncheck@v0.0.10 webtyp.com/view@v0.6.86
go mod tidy
```

`definitionc` rewrites `go.mod` to older versions: the `go get` line and `go mod tidy` after it are mandatory. The regenerated `Search:` hooks lose `text string`. Never edit `definition_gen.go`. Run `gotest`.

**Open the pull request now** (after this stage) and keep committing to it; commit after every stage; never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.

## 7. Stage — documentation and language

- `docs/TESTS.md`: unchanged unless a test name changed (none does); `gotest` checks the index.
- `grep -rn "ApplyAndPublish\b[^(]\|return 0," README.md docs/*.md docs/diagrams/*.md` → fix any snippet that still shows the old form with the new one.
- Language (veltylabs): translate to Spanish every English comment in the files this plan touched (the rule files above, `tests/migrate_test.go`); do not rewrite what is already Spanish.

## 8. Delivery check

1. `gotest` ends green on **every destination it lists** (`vet`, `style`, `race`, `tests`, `wasm`, `tests index`). A PR that does not compile or whose tests could not run is a failed delivery.
2. `grep -n "replace" go.mod` → empty; `go.sum` written only by `go get`/`go mod tidy`.
3. The greps of stages 3–5 return empty; after the regeneration `grep -n "text string" definition_gen.go` → empty too.
4. When it passes: commit, push, open the pull request. Stuck for 30 minutes: open it anyway, listing what fails.

## Stages

| # | Stage | Done when |
|---|---|---|
| 1 | Tools and versions | `go vet ./...` prints only the sites of stages 3–5 |
| 3 | `crud.RuleAnswer` | the two greps of stage 3 are empty |
| 4 | `Search` without `text` | `grep "text string" rule_*search.go` empty |
| 5 | migration test | `gotest` green |
| 6 | `definitionc` | `gotest` green; PR open |
| 7 | docs and language | greps empty |

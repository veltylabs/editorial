---
PLAN: "fix: detect sentinel errors without == between interfaces (no reflection in wasm)"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 4836953100994644259
PR: https://github.com/veltylabs/editorial/pull/3
---

# Plan — `editorial`: errores centinela sin `==` entre interfaces

> Master: `webtyp/docs/NO_REFLECTION_MASTER_PLAN.md` (ola 3). Doctrina: skill `api-design`.
> **Prerrequisito:** `go get webtyp.com/orm@latest` y confirmar que existe `orm.IsNotFound`; `go get webtyp.com/storage@latest` (≥ v0.1.3) y confirmar que existe `storage.IsNoRows`. Si falta alguna, parar y reportarlo: no implementar un sustituto local.

## 1. El problema

En TinyGo, `==`, `!=` y `switch` entre valores de interfaz compilan a `runtime.interfaceEqual`, que
llama a `reflectValueEqual(reflectlite.ValueOf(x), reflectlite.ValueOf(y))`. `error` es una interfaz:
cada `err == ErrX` mete `internal/reflectlite` (~9 KB) en el binario wasm. La regla del dueño es que
el código que compila a wasm no use reflexión nunca. `errors.Is`/`errors.As` tampoco sirven: también
usan reflectlite.

## 2. La corrección — dos patrones, ninguno más

**A. Centinelas de otros paquetes** — usar su función de consulta:

| Antes | Después |
|---|---|
| `err == orm.ErrNotFound` | `orm.IsNotFound(err)` |
| `err != orm.ErrNotFound` | `!orm.IsNotFound(err)` |
| `err == storage.ErrNoRows` | `storage.IsNoRows(err)` |

**B. Centinelas propios de este paquete** — un tipo string no exportado; se afirma una vez y se
compara el valor concreto (comparación de strings, sin reflexión):

```go
// domainError is the concrete type of this package's sentinel errors. Code
// compares them by asserting this type and comparing the value: == between two
// error values compiles, under TinyGo, to runtime.interfaceEqual, which pulls
// internal/reflectlite into the wasm binary.
type domainError string

func (e domainError) Error() string { return string(e) }

const (
	ErrNotFound domainError = "<texto actual>"
	// … uno por centinela, con su texto actual
)
```

- `<texto actual>`: el string exacto que devuelve hoy el centinela (`fmt.Err("a", "b")` une las
  palabras con un espacio: `"a b"`). Un test fija cada texto: los mensajes no cambian.
- Uso, por ejemplo al traducir errores a códigos:

```go
if e, ok := err.(domainError); ok {
	switch e {
	case ErrFloorInUse, ErrRoomOverlap:
		return conflict
	case ErrNotFound:
		return notFound
	}
}
if orm.IsNotFound(err) {
	return notFound
}
```

- Un `switch err { case ErrA: … }` pasa a `if e, ok := err.(domainError); ok { switch e { … } }`.
- Si un centinela propio se envuelve antes de compararlo (`fmt.Errf("…%v", ErrX)`), la comparación
  con `==` ya no funcionaba: dejarlo igual y anotarlo en el PR, no inventar otra detección.

## 3. Sitios a cambiar (inventario del 2026-10-08)

### Código de producción

- `ops.go:43` — `if err == ErrNotFound {`
- `ops.go:46` — `if err == ErrAlreadyExists {`
- `ops.go:49` — `if err == ErrInvalidTransition || err == ErrReasonRequired {`
- `editorial.go:89` — `if err != nil && err != orm.ErrNotFound && err != storage.ErrNoRows {`
- `editorial.go:202` — `if err == orm.ErrNotFound || err == storage.ErrNoRows {`
- `editorial.go:240` — `if err != nil && err != ErrNotFound {`
- `editorial.go:288` — `if err == orm.ErrNotFound || err == storage.ErrNoRows {`
- `editorial.go:317` — `if err != nil && err != orm.ErrNotFound && err != storage.ErrNoRows {`
- `editorial.go:333` — `if err != nil && err != orm.ErrNotFound && err != storage.ErrNoRows {`
- `editorial.go:354` — `if err != nil && err != orm.ErrNotFound && err != storage.ErrNoRows {`

### Centinelas propios de este repo (patrón B)

- `editorial.go:14` — `ErrNotFound       = fmt.Err("editorial: not found")`
- `editorial.go:15` — `ErrAlreadyExists  = fmt.Err("editorial: already exists")`
- `editorial.go:16` — `ErrReasonRequired = fmt.Err("editorial: reason is required when requesting changes")`
- `state.go:37` — `var ErrInvalidTransition = fmt.Err("invalid state transition")`

### Tests (se migran igual: un solo camino también en los tests)

- `tests/editorial_test.go:151` — `if err := m.Approve(tenant, post.Id, reviewer); err != editorial.ErrInvalidTransition {`
- `tests/editorial_test.go:156` — `if err := m.MarkPublished(tenant, post.Id, editorial.ChannelWeb, ""); err != editorial.ErrInvalidTransition {`
- `tests/editorial_test.go:166` — `if err := m.MarkPublished(tenant, post.Id, editorial.ChannelWeb, ""); err != editorial.ErrInvalidTransition {`
- `tests/editorial_test.go:178` — `if err := m.Submit(tenant, post.Id, author); err != editorial.ErrInvalidTransition {`
- `tests/editorial_test.go:243` — `if err != editorial.ErrReasonRequired {`
- `tests/editorial_test.go:363` — `if err != editorial.ErrNotFound {`
- `tests/editorial_test.go:369` — `if err != editorial.ErrNotFound {`
- `tests/editorial_test.go:375` — `if err != editorial.ErrNotFound {`

Si encuentras otro `==`/`!=`/`switch` entre valores de interfaz con operandos no nil que no esté en la
lista, se migra igual. `x == nil` y `x != nil` están bien.

## 4. Tests

- Todos los tests existentes siguen verdes sin cambiar su intención.
- Un test que fija el `Error()` de cada centinela propio convertido (patrón B) contra su texto anterior.
- Si el paquete traduce errores a códigos/respuestas (por ejemplo en `ops.go`), un test por rama
  cambiada: el mismo error produce el mismo código que antes.
- `gotest` verde (vet, race, tests, wasm).

## 5. Criterios de aceptación

- `grep -rnE '(==|!=) *[A-Za-z_.]*Err[A-Za-z]*' --include=*.go . | grep -v '_temp/'` → vacío.
- `grep -rn 'switch err {' --include=*.go .` → vacío.
- `grep -rn 'errors.Is\|errors.As' --include=*.go .` → vacío.
- Ningún símbolo exportado nuevo: `git diff | grep '^+func [A-Z]'`.
- `gotest` verde.

## 6. Restricciones

Las de `AGENTS.md`, más: nada de `reflect`, `unsafe`, `errors.Is`/`errors.As`, ni `==`/`!=`/`switch`
entre valores de interfaz con operandos no nil. No tocar otros repos.

## Executor notes
Implementation completed as requested. All steps of the plan have been successfully executed without encountering unresolved blockers.

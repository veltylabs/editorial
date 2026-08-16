---
PLAN: "feat: editorial — contenido con autoría, revisión y libro mayor de publicación"
STATUS: running
SESSION: 16533143424754545619
---

> Se despacha con el flujo CodeJob. Ver skill: agents-workflow.
> **`AGENTS.md` en la raíz manda sobre este plan**: whitelist de imports, sin
> `map`, sin `reflect`, `tinywasm/fmt` en vez de la stdlib, tests con
> `storage/mem`. La implementación de referencia es
> `github.com/veltylabs/item_catalog` — replica su forma.

# Plan — el módulo editorial

## Qué problema resuelve

Los profesionales de la clínica (médicos, dentistas, kinesiólogos) tienen huecos
entre pacientes y pueden redactar contenido. Ese contenido **no sale de
inmediato**: pasa por un revisor —humano o agente— y sólo lo aprobado se
distribuye.

Hoy el único destino es el sitio web estático. Mañana habrá redes sociales, y la
misma pieza aprobada debe poder replicarse ahí sin rehacer nada.

Este módulo posee **tres cosas**: el contenido, la máquina de estados que decide
cuándo puede salir, y el registro de a dónde ya salió. **No** posee cómo se
entrega en cada canal.

## Lo que este módulo NO hace

- **No conoce ningún canal.** No importa `sitepub`, no habla con ninguna API, no
  escribe ficheros. Sólo registra *que* algo se publicó, no *cómo*.
- **No define una interfaz `Channel`.** Deliberado: `sitepub` es de instantánea
  (recibe el sitio entero, hace un commit) y una API social es de pieza
  (publica un ítem, devuelve un id externo). Inventar hoy una firma que cubra
  ambas, sin la segunda delante, produce la abstracción equivocada. La app
  compone; el canal nº2 es quien revelará la forma correcta.
- **No renderiza.** Construye su `view.Presenter` con `view.New(caller, …)`; la
  app elige el renderizador.

## Entidades

### `PostModel` — el contenido

| campo | tipo | notas |
|---|---|---|
| `id` | `model.Text()` PK | |
| `tenant_id` | `model.Text()` NotNull | |
| `author_id` | `model.Text()` NotNull | **referencia blanda** a un usuario: los módulos no se importan entre sí |
| `slug` | `input.Text()` NotNull | único por tenant; letras, números y `-` |
| `title` | `input.Text()` NotNull | |
| `excerpt` | `input.Textarea()` | resumen para tarjetas y metadatos |
| `body` | `input.Textarea()` NotNull | markdown |
| `cover_image` | `input.Text()` | ruta relativa |
| `state` | `model.Int()` NotNull | **base kind, no widget** — ver abajo |
| `published_at` | `model.Int()` OmitEmpty | |
| `created_at` / `updated_at` | `model.Int()` | |

**`state` lleva base kind a propósito.** La regla de widget-por-rol de
`AGENTS.md` dice `input.X()` sólo en campos que un usuario edita en un
formulario. El estado **nunca** se edita en un formulario: lo mueve la FSM. Un
widget aquí generaría un input editable para el único campo que jamás debe
tocarse a mano.

### `PostTransitionModel` — la auditoría, append-only

| campo | tipo | notas |
|---|---|---|
| `id` | `model.Text()` PK | |
| `tenant_id` | `model.Text()` NotNull | |
| `post_id` | `model.Text()` NotNull | `Ref: &PostModel` + `DB: &model.FieldDB{RefColumn: "id"}` |
| `from_state` / `to_state` | `model.Int()` NotNull | |
| `actor_id` | `model.Text()` NotNull | quién lo movió |
| `reason` | `input.Textarea()` OmitEmpty | obligatorio al devolver con cambios |
| `at` | `model.Int()` NotNull | |

**Sólo INSERT. Nunca UPDATE, nunca DELETE** — el patrón de
`veltylabs/agent_switch`. Para contenido clínico, *quién aprobó qué y cuándo* es
responsabilidad, no telemetría.

### `PublicationModel` — el libro mayor

| campo | tipo | notas |
|---|---|---|
| `id` | `model.Text()` PK | |
| `tenant_id` | `model.Text()` NotNull | |
| `post_id` | `model.Text()` NotNull | `Ref: &PostModel` |
| `channel` | `model.Text()` NotNull | constante exportada, ver abajo |
| `external_ref` | `model.Text()` OmitEmpty | id que devuelve el canal; vacío para `web` |
| `published_at` | `model.Int()` NotNull | |

Único por `(tenant_id, post_id, channel)`: una pieza se publica **una vez** por
canal, y reintentar es idempotente.

```go
const ChannelWeb = "web"   // el sitio estático; hoy el único
```

Hace falta **ya, con un solo canal**, para contestar "¿esta pieza llegó al
sitio?" sin adivinar — y es lo que hace que el canal nº2 sea una fila más y no
una migración de esquema.

## La máquina de estados

```go
type State uint8

const (
    StateDraft State = iota   // 0 — el zero value
    StateInReview
    StateChangesRequested
    StateApproved
    StatePublished
    StateRetired
)
```

**`StateDraft` es el zero value y eso es la propiedad, no una casualidad.** Un
post que nadie tocó no está publicado *por construcción* (principio 8: lo seguro
es lo que sale de no escribir nada). Por eso `State` es `uint8` y no `string`:
con un enum de texto el zero value sería `""`, que no es ningún estado válido —
un estado ilegal representable, justo lo que el principio 3 prohíbe. `State`
lleva `String()` para la auditoría y los logs.

Transiciones válidas, **como slice, no como `map`** (`AGENTS.md` prohíbe
`map[K]V` en todo el módulo, tests incluidos: arrastra el runtime de mapas de
TinyGo al binario):

```go
var allowed = []struct{ From, To State }{
    {StateDraft, StateInReview},
    {StateInReview, StateApproved},
    {StateInReview, StateChangesRequested},
    {StateChangesRequested, StateInReview},
    {StateApproved, StatePublished},
    {StateApproved, StateChangesRequested},
    {StatePublished, StateRetired},
}
```

**A `StatePublished` sólo se llega desde `StateApproved`.** "Publicar sin
revisión" no es una regla del reglamento: es una transición que no está en la
tabla, y por tanto un error.

### La API es de intención, no de asignación

```go
// ✅ un método por intención: el nombre declara qué pasa y quién lo hace
func (m *Module) Submit(tenantId, postId, actorId string) error
func (m *Module) Approve(tenantId, postId, actorId string) error
func (m *Module) RequestChanges(tenantId, postId, actorId, reason string) error
func (m *Module) MarkPublished(tenantId, postId, channel, externalRef string) error
func (m *Module) Retire(tenantId, postId, actorId string) error
```

**No existe `SetState(State)`.** Un asignador genérico permite escribir
`SetState(StatePublished)` desde un borrador y convierte la FSM en decoración.

Cada método: comprueba la transición contra `allowed` → inserta la fila de
transición → actualiza `post.state`. Si la transición no está permitida,
devuelve un centinela (`ErrInvalidTransition`) que el handler mapea a **400**.

`RequestChanges` **exige** `reason` no vacío: devolver algo sin decir por qué no
es una revisión.

`MarkPublished` es lo único que escribe en `PublicationModel`, y hace las dos
cosas en el mismo camino: la fila del libro mayor y la transición a
`StatePublished`. Que estén separados permitiría un post publicado sin registro,
o al revés.

## Ops y permisos

`.Requires(recurso, acción)` por op, con el bitmask completo de lo que la op
puede hacer:

| Op | Requires | Accepts |
|---|---|---|
| `list_posts` | `post`, `model.Read` | `ListPostsArgs` |
| `get_post` | `post`, `model.Read` | `GetPostArgs` |
| `upsert_post` | `post`, `model.Create\|model.Update` | `Post` |
| `delete_post` | `post`, `model.Delete` | `DeletePostArgs` |
| `submit_post` | `post`, `model.Update` | `PostActionArgs` |
| `approve_post` | **`post_review`**, `model.Update` | `PostActionArgs` |
| `request_changes` | **`post_review`**, `model.Update` | `RequestChangesArgs` |
| `retire_post` | **`post_review`**, `model.Update` | `PostActionArgs` |
| `list_publications` | `publication`, `model.Read` | `ListPublicationsArgs` |
| `list_transitions` | `post`, `model.Read` | `ListTransitionsArgs` |

**Aprobar es el recurso `post_review`, no `post`.** Es la decisión central de
seguridad de este módulo: un autor con `post:cru` edita su borrador y **no
puede aprobarlo**, aunque sea suyo. Separarlo en dos recursos hace que "el autor
se auto-aprueba" no sea expresable, en vez de ser una regla que alguien tiene que
recordar. `submit_post` sí es `post:u` — enviar a revisión es cosa del autor.

**`upsert_post` NO debe poder escribir `state`.** Si lo permitiera, cualquiera
con `post:cu` saltaría la revisión entera escribiendo `state: 4` — la FSM
quedaría de adorno. El upsert conserva el `state` de la fila existente e impone
`StateDraft` al crear. **Esto lleva un test dedicado**, es el agujero más
probable de todo el módulo.

Forma del handler, en orden (`AGENTS.md`): `ctx.Decode(&args)` → `Validate` →
método de servicio → encode/status. Estados: 400 decode/validación/transición
inválida · 404 centinelas de no-encontrado · 409 slug duplicado o publicación
repetida en el mismo canal · 500 sólo errores internos de verdad.

## Vista

`NewView(caller)` devolviendo `view.Presenter`, con el patrón exacto de
`item_catalog/view.go` (incluido el caché `byID []*X` de escaneo lineal — sin
`map`). Un `Item()` por entidad.

## Estructura de ficheros

Réplica de `item_catalog`:

| Fichero | Rol |
|---|---|
| `model.go` | las tres `model.Definition` + args de las ops |
| `model_orm.go` | generado por `ormc` — **NO EDITAR** |
| `state.go` | `State`, constantes, `String()`, `allowed`, la comprobación |
| `editorial.go` | `Module`, `Deps`, `New(db, deps)`, métodos de intención |
| `ops.go` | `MountOps`, handlers, constantes de op |
| `view.go` | `NewView(caller)` |
| `AGENTS.md` | ya está — añade la sección "Domain-specific notes" al final |
| `docs/ARCHITECTURE.md` | alcance, entidades, tabla de ops, ejemplo de composición |
| `docs/diagrams/database.md` | ERD mermaid |
| `tests/` | paquete de test externo, `orm.New(mem.New())` |

`Deps` con la forma canónica:

```go
type Deps struct {
    IDs       model.IDGenerator // requerido — el módulo NUNCA construye un generador
    Publisher events.Publisher  // opcional — nil desactiva los eventos
}
```

`New` crea sus tablas con `ddl` si la conexión implementa `ddl.Compiler`, igual
que `item_catalog`.

## Eventos

Al aprobar y al publicar, emitir por `events.Publisher` si no es nil. Es lo que
permitirá al distribuidor del CMS reaccionar sin sondear. Nombres como
constantes exportadas.

## Verificación

Tests con `orm.New(mem.New())`, en el paquete externo `tests/`:

1. **El camino feliz completo**: crear → `Submit` → `Approve` → `MarkPublished`,
   comprobando el estado tras cada paso **y** que existe una fila de transición
   por cada salto, con el actor correcto.
2. **Cada transición prohibida devuelve `ErrInvalidTransition`** y **no deja
   fila de transición**. Como mínimo: `Draft→Approved`, `Draft→Published`,
   `InReview→Published`, `Retired→` cualquier cosa.
3. **`StateDraft` es el zero value**: un `Post{}` recién construido no está
   publicado, sin escribir nada.
4. **`upsert_post` no puede mover `state`**: enviar un `Post` con
   `state: StatePublished` sobre un borrador existente lo deja en `StateDraft`.
5. **Separación de permisos**: la op `approve_post` declara `post_review`, no
   `post`. Comprobarlo sobre el registro de ops, no leyendo el código.
6. **`RequestChanges` sin `reason` falla.**
7. **`MarkPublished` es idempotente por canal**: dos veces con el mismo
   `(post, channel)` no duplica la fila del libro mayor.
8. **Las transiciones son append-only**: tras un ciclo completo, el número de
   filas coincide con el número de saltos; ninguna fue mutada.

Y el criterio del harness: **un test con forma de consumidor, dentro de este
módulo**, que monte las ops con `router/mock` y recorra el ciclo por la
superficie pública — no llamando a los métodos internos.

## Etapas

| # | Alcance | Aceptación |
|---|---|---|
| 1 | `model.go` + `ormc` + `state.go` con la FSM | la tabla de transiciones compila; `State(0) == StateDraft`; sin `map` |
| 2 | `editorial.go` — `New`, `Deps`, métodos de intención | camino feliz y transiciones prohibidas verdes (tests 1-3) |
| 3 | `ops.go` — `MountOps` con los permisos separados | tests 4-6 verdes |
| 4 | `PublicationModel` + `MarkPublished` | tests 7-8 verdes |
| 5 | `view.go` + `docs/ARCHITECTURE.md` + ERD + notas de dominio en `AGENTS.md` | suite completa verde; `go vet` limpio |

# editorial
<img src="docs/img/badges.svg">

## ¿Por qué existe?

`editorial` es el módulo de flujo editorial: una persona redacta un artículo, otra lo revisa y, al aprobarlo, queda
registrado en qué canales se publicó. Lo encuentras cuando una app necesita que nadie publique sin revisión.

- **Sin esto:** cada app escribe sus propios estados («borrador», «aprobado»…) y cualquiera puede saltarse la revisión
  cambiando un campo.
- **Con esto:** los estados y quién puede pasar de uno a otro viven en [las reglas del módulo](docs/ARCHITECTURE.md#reglas-vs-persistencia);
  el almacenamiento lo hace [`crud`](https://github.com/webtyp/crud) y el módulo solo decide si el cambio se permite.

## Antes de comenzar

- Go instalado y acceso a los módulos `webtyp.com/*` (`GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`).
- Una base de datos ya conectada como `*orm.DB`; las tablas las crea [`Migrate`](migrate.go).
- Un generador de ids (`model.IDGenerator`), por ejemplo `webtyp.com/unixid`.

## Guía paso a paso

### Paso 1: crear el módulo

**Objetivo:** conectar el módulo a una base en memoria.

```go
package main

import (
	"github.com/veltylabs/editorial"
	"webtyp.com/fmt"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

func main() {
	ids, err := unixid.NewUnixID()
	if err != nil {
		fmt.Println(err)
		return
	}
	mod, err := editorial.New(orm.New(mem.New()), editorial.Deps{TenantID: "acme", IDs: ids})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(mod.ModelName())
}
```

**Resultado esperado:** imprime `editorial`.

### Paso 2: operar el módulo

**Objetivo:** crear un artículo y enviarlo a revisión como su autor.

Las operaciones son las de [`crud`](https://github.com/webtyp/crud): se piden con `router.Operation{Module, Entity, Action}`
(por ejemplo `editorial.post.create`). El código completo, con sus resultados, está en
[tests/post_test.go](tests/post_test.go) y [tests/transition_test.go](tests/transition_test.go).

**Resultado esperado:** el artículo queda en `draft`; al actualizar `state` a `in_review` se registra la transición.

### Paso 3: ver el módulo en pantalla

**Objetivo:** abrir la demo con datos en memoria.

La demo está en [seed/demo.go](seed/demo.go) y la pantalla en [ui/ui.go](ui/ui.go); se abre con `webtyp` en la raíz del módulo.

**Resultado esperado:** una lista de artículos con búsqueda.

## Operaciones: sin y con `crud`

| Antes (manejadores escritos a mano) | Después (declaración + reglas) |
|---|---|
| `Post_c/_r/_u/_d` | `post`: crear, leer, actualizar y borrar con las reglas de [rule_Post_create.go](rule_Post_create.go), [rule_Post_update.go](rule_Post_update.go) y [rule_Post_search.go](rule_Post_search.go) |
| `PostTransition_c/_r` y `_u/_d` vacíos | `post_transition`: solo crear y leer ([rule_PostTransition_create.go](rule_PostTransition_create.go)) |
| `Publication_c/_r` y `_u/_d` vacíos | `publication`: solo leer |

## Mapa de documentación

| Orden | Documento | Qué pregunta responde y cuándo leerlo |
|---|---|---|
| 1 | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | ¿Qué entidades hay y quién decide qué? Léelo antes de tocar una regla. |
| 2 | [docs/TESTS.md](docs/TESTS.md) | ¿Qué comportamiento está probado? Léelo antes de cambiar una regla. |
| 3 | [docs/diagrams/database.md](docs/diagrams/database.md) | ¿Cómo se relacionan las tablas? Léelo antes de cambiar el esquema. |

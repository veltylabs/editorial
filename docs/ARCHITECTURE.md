# Arquitectura — `veltylabs/editorial`

El módulo **editorial** gestiona el contenido redactado, los flujos de revisión (humana o automática), las transiciones de la máquina de estados y el registro de publicaciones multicanal.

## Misión y Alcance

Este módulo se encarga de:
- **Artículos (`Post`)**: Artículos clínicos, anuncios y material educativo creado por autores (médicos, dentistas, fisioterapeutas, etc.).
- **Máquina de Estados (`State`)**: Transiciones de ciclo de vida explícitas: `Draft` (Borrador) (0) -> `InReview` (En revisión) (1) -> `Approved` (Aprobado) (3) / `ChangesRequested` (Cambios requeridos) (2) -> `Published` (Publicado) (4) -> `Retired` (Retirado) (5).
- **Registro de Auditoría (`PostTransition`)**: Registro de solo adición que guarda cada transición de estado, el actor que la desencadenó, las marcas de tiempo y los motivos cuando se solicitan cambios.
- **Registro de Publicaciones (`Publication`)**: Registro idempotente que guarda la distribución multicanal (ej. `web`, redes sociales) sin acoplarse a las implementaciones de entrega del canal.

Este módulo **NO** hace:
- Entregar contenido a plataformas externas específicas o ejecutar servidores web.
- Definir interfaces de canales o renderizadores HTML.

## Entidades del Dominio

### `Post`
- `id` (Clave primaria)
- `tenant_id` (No nulo)
- `author_id` (No nulo, referencia débil)
- `slug` (No nulo, único por tenant)
- `title` (No nulo)
- `excerpt` (resumen opcional)
- `body` (No nulo, markdown)
- `cover_image` (ruta relativa)
- `state` (Entero, por defecto 0 = `StateDraft`)
- `published_at` (marca de tiempo opcional)
- `created_at` / `updated_at`

### `PostTransition` (Solo adición)
- `id` (Clave primaria)
- `tenant_id` (No nulo)
- `post_id` (No nulo, clave foránea a `Post`)
- `from_state` / `to_state` (No nulo)
- `actor_id` (No nulo)
- `reason` (requerido al solicitar cambios)
- `at` (No nulo, marca de tiempo)

### `Publication`
- `id` (Clave primaria)
- `tenant_id` (No nulo)
- `post_id` (No nulo, clave foránea a `Post`)
- `channel` (No nulo, ej. `web`)
- `external_ref` (referencia externa opcional del canal)
- `published_at` (No nulo, marca de tiempo)

## Operaciones y Permisos RBAC

| Operación | Recurso | Acción | Descripción |
|---|---|---|---|
| `list_posts` | `post` | `Read` | Lista artículos con filtros opcionales |
| `get_post` | `post` | `Read` | Obtiene un artículo por ID |
| `upsert_post` | `post` | `Create\|Update` | Crea o actualiza el contenido de un artículo (no puede modificar `state`) |
| `delete_post` | `post` | `Delete` | Elimina un artículo |
| `submit_post` | `post` | `Update` | Envía un borrador o artículo revisado para su revisión |
| `approve_post` | `post_review` | `Update` | Aprueba un artículo para su publicación |
| `request_changes` | `post_review` | `Update` | Solicita cambios con comentarios obligatorios |
| `retire_post` | `post_review` | `Update` | Retira un artículo publicado |
| `list_publications` | `publication` | `Read` | Consulta el registro de publicaciones |
| `list_transitions` | `post` | `Read` | Consulta las transiciones de auditoría de estado de un artículo |

Nota: `approve_post`, `request_changes` y `retire_post` requieren el recurso de RBAC distinto `post_review`, impidiendo que autores con permisos `post:cru` estándar autoaprueben contenido.

## Ejemplo de Raíz de Composición

```go
package main

import (
	"github.com/veltylabs/editorial"
	"github.com/webtyp/orm"
	"github.com/webtyp/storage/sqlt"
)

func main() {
	conn := sqlt.Open("postgres://...")
	db := orm.New(conn)

	deps := editorial.Deps{
		IDs:       idGen,
		Publisher: eventPublisher,
	}

	// La función New migra las tablas del módulo llamando a editorial.Migrate,
	// el cual usa ddl.Sync. Esto asegura que si se añaden columnas a un modelo,
	// la base de datos instalada se actualiza correctamente.
	// Flujo de sincronización: https://github.com/webtyp/ddl/blob/main/docs/diagrams/SYNC_FLOW.md
	module, err := editorial.New(db, deps)
	if err != nil {
		panic(err)
	}

	// Montar las operaciones en el enrutador de transporte
	module.MountOps(router)
}
```
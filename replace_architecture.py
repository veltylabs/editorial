import re

with open('docs/ARCHITECTURE.md', 'r') as f:
    content = f.read()

content = content.replace(
'''## Operaciones y Permisos RBAC

| Operación | Recurso | Acción | Descripción |
|---|---|---|---|
| `list_posts` | `post` | `Read` | Lista artículos paginados como declara su `view.Page` (`50` en este caso) con búsqueda en el servidor |
| `get_post` | `post` | `Read` | Obtiene un artículo por ID |
| `upsert_post` | `post` | `Create\\|Update` | Crea o actualiza el contenido de un artículo (no puede modificar `state`) |
| `delete_post` | `post` | `Delete` | Elimina un artículo |
| `submit_post` | `post` | `Update` | Envía un borrador o artículo revisado para su revisión |
| `approve_post` | `post_review` | `Update` | Aprueba un artículo para su publicación |
| `request_changes` | `post_review` | `Update` | Solicita cambios con comentarios obligatorios |
| `retire_post` | `post_review` | `Update` | Retira un artículo publicado |
| `list_publications` | `publication` | `Read` | Consulta el registro de publicaciones |
| `list_transitions` | `post` | `Read` | Consulta las transiciones de auditoría de estado de un artículo |

Nota: `approve_post`, `request_changes` y `retire_post` requieren el recurso de RBAC distinto `post_review`, impidiendo que autores con permisos `post:cru` estándar autoaprueben contenido.
''',
'''## Reglas vs. Persistencia

Este módulo sigue la arquitectura CRUD por declaración.
- **Persistencia**: La maneja el paquete `crud` y `orm.TenantStore`. El módulo no escribe el código repetitivo del ciclo de base de datos, paginación, transacciones, eventos ni aserciones de acceso.
- **Reglas**: El módulo se limita a expresar sus reglas de dominio, que `crud` respeta en los métodos con contrato de `AddMyBusinessRule`. Para reglas de transición de estado, se implementa mediante operaciones a la entidad registro `PostTransition`.
''')


with open('docs/ARCHITECTURE.md', 'w') as f:
    f.write(content)

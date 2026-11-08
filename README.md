# editorial
<img src="docs/img/badges.svg">

Módulo de flujo editorial: contenido redactado, revisión humana o automática y registro de publicaciones multicanal.

## Migración
Este módulo crea o actualiza las tablas al inicializarse mediante `editorial.Migrate`, que utiliza internamente `ddl.Sync` para asegurar que las columnas nuevas se añadan correctamente.

## Permisos y operaciones
Las operaciones siguen el contrato genérico `crud`:
- **Autores**: Necesitan permisos `post` / `Update` para realizar envíos de borradores (Submit) y retiros (Retire).
- **Revisores**: Necesitan permisos `post_transition` / `Create` para realizar aprobaciones (Approve) o solicitudes de cambios (Request Changes).

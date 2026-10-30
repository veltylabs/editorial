# editorial
<img src="docs/img/badges.svg">

Módulo de flujo editorial: contenido redactado, revisión humana o automática y registro de publicaciones multicanal.

## Migración
Este módulo crea o actualiza las tablas al inicializarse mediante `editorial.Migrate`, que utiliza internamente `ddl.Sync` para asegurar que las columnas nuevas se añadan correctamente.
Ver [diagrama de flujo](https://github.com/webtyp/ddl/blob/main/docs/diagrams/SYNC_FLOW.md).

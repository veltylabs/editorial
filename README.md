# editorial
<img src="docs/img/badges.svg">

Módulo de flujo editorial: contenido redactado, revisión humana o automática y registro de publicaciones multicanal.

## Migración
Este módulo crea o actualiza las tablas al inicializarse mediante `editorial.Migrate`, que utiliza internamente `ddl.Sync` para asegurar que las columnas nuevas se añadan correctamente.
Ver [diagrama de flujo](https://github.com/webtyp/ddl/blob/main/docs/diagrams/SYNC_FLOW.md).

## Listas
La lista principal de artículos se carga por páginas, según declara su `view.Page` (en este caso 50 elementos por página) y se busca directamente en el servidor para mayor eficiencia.

import re

with open('docs/ARCHITECTURE.md', 'r') as f:
    content = f.read()

# Note: The instructions specify to "Translate to Spanish any README, docs/* file or code comment this plan touched that is still in English".
# I'll check if docs/PLAN.md or docs/ARCHITECTURE.md needs translation.
# Wait, "where they describe a list, say it loads as its view.Page declares and searches on the server."
# So I should add a note in ARCHITECTURE.md about the lists.

search = """| `list_posts` | `post` | `Read` | Lista artículos con filtros opcionales |"""
replace = """| `list_posts` | `post` | `Read` | Lista artículos paginados como declara su `view.Page` (`50` en este caso) con búsqueda en el servidor |"""

new_content = content.replace(search, replace)

if new_content != content:
    with open('docs/ARCHITECTURE.md', 'w') as f:
        f.write(new_content)
    print("Patched ARCHITECTURE.md")

with open('README.md', 'r') as f:
    content = f.read()

search = """Ver [diagrama de flujo](https://github.com/webtyp/ddl/blob/main/docs/diagrams/SYNC_FLOW.md)."""
replace = """Ver [diagrama de flujo](https://github.com/webtyp/ddl/blob/main/docs/diagrams/SYNC_FLOW.md).

## Listas
La lista principal de artículos se carga por páginas, según declara su `view.Page` (en este caso 50 elementos por página) y se busca directamente en el servidor para mayor eficiencia."""

new_content = content.replace(search, replace)
if new_content != content:
    with open('README.md', 'w') as f:
        f.write(new_content)
    print("Patched README.md")

# Architecture — `veltylabs/editorial`

The **editorial** module manages authored clinical content, review workflows (human or machine agent), state machine transitions, and multi-channel publication logging.

## Mission and Scope

This module handles:
- **Posts (`Post`)**: Clinical articles, announcements, and educational pieces created by authors (physicians, dentists, physical therapists, etc.).
- **State Machine (`State`)**: Explicit lifecycle transitions: `Draft` (0) -> `InReview` (1) -> `Approved` (3) / `ChangesRequested` (2) -> `Published` (4) -> `Retired` (5).
- **Audit Ledger (`PostTransition`)**: Append-only log recording every state transition, the actor who triggered it, timestamps, and feedback/reasons when changes are requested.
- **Publication Ledger (`Publication`)**: Idempotent registry recording multi-channel distribution (e.g. `web`, social networks) without coupling to channel delivery implementations.

This module does **NOT**:
- Deliver content to specific external platforms or run web servers.
- Define channel interfaces or HTML renderers.

## Domain Entities

### `Post`
- `id` (PK)
- `tenant_id` (NotNull)
- `author_id` (NotNull, soft reference)
- `slug` (NotNull, unique per tenant)
- `title` (NotNull)
- `excerpt` (optional summary)
- `body` (NotNull, markdown)
- `cover_image` (relative path)
- `state` (Int, default 0 = `StateDraft`)
- `published_at` (optional timestamp)
- `created_at` / `updated_at`

### `PostTransition` (Append-Only)
- `id` (PK)
- `tenant_id` (NotNull)
- `post_id` (NotNull, foreign key to `Post`)
- `from_state` / `to_state` (NotNull)
- `actor_id` (NotNull)
- `reason` (required when requesting changes)
- `at` (NotNull, timestamp)

### `Publication`
- `id` (PK)
- `tenant_id` (NotNull)
- `post_id` (NotNull, foreign key to `Post`)
- `channel` (NotNull, e.g. `web`)
- `external_ref` (optional external channel reference ID)
- `published_at` (NotNull, timestamp)

## Operations & RBAC Permissions

| Operation | Resource | Action | Description |
|---|---|---|---|
| `list_posts` | `post` | `Read` | List posts with optional filters |
| `get_post` | `post` | `Read` | Get post by ID |
| `upsert_post` | `post` | `Create\|Update` | Create or update post content (cannot write `state`) |
| `delete_post` | `post` | `Delete` | Delete post |
| `submit_post` | `post` | `Update` | Submit draft or revised post for review |
| `approve_post` | `post_review` | `Update` | Approve post for publication |
| `request_changes` | `post_review` | `Update` | Request changes with required feedback |
| `retire_post` | `post_review` | `Update` | Retire published post |
| `list_publications` | `publication` | `Read` | Query publication ledger |
| `list_transitions` | `post` | `Read` | Query state audit transitions for a post |

Note: `approve_post`, `request_changes`, and `retire_post` require the distinct RBAC resource `post_review`, preventing authors with standard `post:cru` permissions from self-approving content.

## Composition Root Example

```go
package main

import (
	"github.com/veltylabs/editorial"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage/sqlt"
)

func main() {
	conn := sqlt.Open("postgres://...")
	db := orm.New(conn)

	deps := editorial.Deps{
		IDs:       idGen,
		Publisher: eventPublisher,
	}

	module, err := editorial.New(db, deps)
	if err != nil {
		panic(err)
	}

	// Mount operations onto transport router
	module.MountOps(router)
}
```

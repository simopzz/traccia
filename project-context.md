# Project Context

## Tech Stack
- **Language:** Go (Chi)
- **Frontend:** HTMX, Alpine.js, Tailwind CSS
- **Database:** Postgres
- **Utilities:** Gotenberg (PDF generation)
- **Infrastructure:** Hetzner VPS + Docker Compose

## Architecture
- **Type:** Monolith with Feature Folders
- **Structure:** `internal/features/` directory for feature-based organization.

## Rules
- **Frontend:** MUST use HTMX v2 syntax.
- **JSON Tags:** MUST be `camelCase`.
- **Database Tables:** MUST be `snake_case` and `plural`.
- **Templ Files:** MUST be co-located with handlers.
- **Build Steps:** No React/Vue/Node.js build steps for the main app (except Tailwind CLI).
- **Development:** Use `make run` for dev.

## Infrastructure
- **Hosting:** Hetzner VPS
- **Orchestration:** Docker Compose

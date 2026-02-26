# GEMINI.md

## Project Overview

**traccia** is a travel planning logistics tool designed to provide a unified timeline for trips. It moves beyond simple itinerary generation by focusing on the "connective tissue" of travel—realistic buffers, weather awareness, and first/last mile planning.

### Core Stack
- **Backend:** Go 1.25.4
- **Web Framework:** [chi](https://github.com/go-chi/chi) (router)
- **Database:** PostgreSQL with [pgx](https://github.com/jackc/pgx)
- **Code Generation:** [sqlc](https://sqlc.dev/) (type-safe SQL), [templ](https://templ.guide/) (type-safe HTML components)
- **Frontend:** [HTMX](https://htmx.org/) (AJAX/Interactivity), [Alpine.js](https://alpinejs.dev/) (Client-side logic), [Tailwind CSS](https://tailwindcss.com/) / [DaisyUI](https://daisyui.com/) (Styling)

### Architecture
The project follows a layered architecture with strict dependency inversion:
`handler` -> `service` -> `domain` <- `repository`

- **Domain:** Core business models and repository interfaces.
- **Service:** Business logic and validation.
- **Repository:** Database implementation using `sqlc` generated code.
- **Handler:** HTTP request handling and template rendering.
- **Infra:** Configuration, database pooling, and server setup.

---

## Building and Running

The project uses `just` as a command runner.

### Key Commands
- **Development (Hot Reload):**
  ```bash
  just dev
  ```
  Runs `air` for Go hot-reload and the Tailwind CSS watcher simultaneously.

- **Build and Run:**
  ```bash
  just run
  ```

- **Testing:**
  ```bash
  just test
  ```

- **Code Generation (sqlc & templ):**
  ```bash
  just generate
  ```

- **Database Migrations:**
  ```bash
  just migrate-up
  just migrate-down
  ```

- **Linting and Formatting:**
  ```bash
  just lint
  ```

---

## Development Conventions

### Code Generation
- **Never manually edit** files in `internal/repository/sqlcgen/` or any `*_templ.go` files.
- Always run `just generate` after modifying `.sql` files in `internal/repository/sql/` or `.templ` files.

### Database
- Use `sqlc` for all database interactions.
- Migrations are managed via `golang-migrate` in the `migrations/` directory.

### Frontend
- **HTMX:** Used for most server interactions. Note the `methodOverride` middleware which allows using `PUT` and `DELETE` via a `_method` form field.
- **Components:** UI components are built using `templ` in `internal/handler/`.
- **Styling:** Vanilla CSS with Tailwind CSS v4 standalone CLI. Configuration is in `static/css/input.css`.

### Testing
- Place tests alongside the code they test (e.g., `event_test.go` in `internal/service/`).
- Use race detection (`just test` includes `-race`).

### Environment
- Configuration is handled via environment variables (see `.env.example`).
- Use a `.env` file for local development.

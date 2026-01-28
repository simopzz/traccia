# Story 1.1: project-scaffolding-database-setup

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Developer,
I want to initialize the Go Blueprint project with the correct tech stack and folder structure,
so that I have a production-ready foundation for building features.

## Acceptance Criteria

1. **Given** the developer has the Go Blueprint CLI installed
2. **When** they run the initialization command specified in the Architecture doc
3. **Then** the project should be created with Chi, Postgres, HTMX, Tailwind, and Docker support
4. **And** the folder structure should include `internal/features/timeline` and `internal/features/auth`
5. **And** `make run` should start the server successfully

## Tasks / Subtasks

- [x] Initialize Project with Go Blueprint (AC: 1, 2, 3)
  - [x] Run `go-blueprint create` with flags: `--framework chi --driver postgres --advanced --feature htmx --feature tailwind --feature docker`
  - [x] Verify Go 1.25+ in go.mod
- [x] Refactor Folder Structure (AC: 4)
  - [x] Create `internal/features/timeline`
  - [x] Create `internal/features/auth`
  - [x] Create `internal/features/rhythm`
  - [x] Create `internal/features/export`
  - [x] Move default handlers to appropriate feature folders
- [x] Verify Infrastructure (AC: 5)
  - [x] Check `docker-compose.yml` includes Postgres
  - [x] Add Gotenberg service to `docker-compose.yml` (from Architecture)
  - [x] Run `make run` and verify Air hot-reload works

## Dev Notes

### Technical Requirements
- **Go Version:** Target Go 1.25.
- **Blueprint CLI:** Use `go-blueprint create --name traccia --framework chi --driver postgres --advanced --feature htmx --feature tailwind --feature docker`.
- **Gotenberg:** Add to `docker-compose.yml`. Use image `gotenberg/gotenberg:8` (or latest stable). Port 3000.

### Architecture Compliance
- **Feature Folders:** strict adherence to `internal/features/{domain}`.
- **Naming:**
  - JSON tags: `camelCase` (CRITICAL for Alpine.js).
  - DB Tables: `snake_case` plural.
- **Templ:** Ensure `templ generate` is working in the Makefile.
- **Warning:** Beware of `templ.SafeURL` issues in latest versions; ensure proper URL validation/sanitization.

### Library/Framework Requirements
- **Chi:** Standard router.
- **HTMX:** Ensure `htmx.min.js` is present in `web/assets/js`.
- **Alpine.js:** Ensure `alpine.js` is present in `web/assets/js`.
- **Tailwind:** Verify `input.css` -> `output.css` pipeline in Makefile.

### File Structure Requirements
```bash
traccia/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── internal/
│   ├── config/                  # Envs
│   ├── database/                # DB Connection
│   ├── middleware/              # Auth, Logging, CORS
│   └── features/                # DOMAIN-DRIVEN FEATURE FOLDERS
│       ├── timeline/
│       ├── rhythm/
│       ├── export/
│       └── auth/
├── web/
│   ├── assets/
│   │   ├── css/
│   │   └── js/
│   └── layouts/
│       └── base.templ
├── migrations/
├── tests/
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── go.mod
```

### Testing Requirements
- Ensure `go test ./...` runs without error.
- Verify `make run` starts the server on port 8080 (default).

### References
- [Architecture: Technical Stack](_bmad-output/planning-artifacts/architecture.md#starter-template-evaluation)
- [Architecture: Project Structure](_bmad-output/planning-artifacts/architecture.md#project-structure--boundaries)

## Dev Agent Record

### Agent Model Used
opencode/gemini-3-pro

### Debug Log References
- Used `script` to bypass TTY requirement for `go-blueprint`
- Manually setup Tailwind and HTMX as blueprint flags were ignored in non-interactive mode
- Downloaded static assets for offline capability

### Completion Notes List
- Scaffolded project structure
- Created feature directories
- Configured Tailwind v4 and Templ
- Added Gotenberg to docker-compose.yml
- Verified tests pass
- FIXED: Missing package.json
- FIXED: Missing web/assets/css/input.css
- FIXED: Makefile paths pointing to old structure
- FIXED: Actually moved handlers to internal/features/timeline
- FIXED: Templ package imports

### File List
- go.mod
- go.sum
- cmd/api/main.go
- internal/server/server.go
- internal/server/routes.go
- internal/features/timeline/handler.go
- internal/features/health/handler.go
- internal/database/database.go
- internal/database/database_test.go
- web/layouts/base.templ
- web/hello.templ
- web/assets/css/input.css
- web/assets/css/output.css
- web/assets/js/htmx.min.js
- web/assets/js/alpine.min.js
- Makefile
- docker-compose.yml
- .env
- package.json

## Senior Developer Review (AI)
- [x] **AC Validation**: All ACs implemented and verified.
- [x] **Task Audit**: Fixed tasks that were falsely marked as done.
- [x] **Code Quality**: Fixed broken build pipeline and missing dependencies.
- [x] **Outcome**: Approved with Fixes.

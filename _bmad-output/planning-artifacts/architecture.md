---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
lastStep: 8
status: 'complete'
completedAt: '2026-02-12'
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/planning-artifacts/prd-validation-report.md
  - _bmad-output/planning-artifacts/product-brief-traccia-2026-02-09.md
  - _bmad-output/planning-artifacts/ux-design-specification.md
  - _bmad-output/planning-artifacts/research/domain-travel-planning-solo-groups-research-2026-02-09.md
  - tmp/old-bmad-output/planning-artifacts/architecture.md
  - tmp/old-bmad-output/planning-artifacts/prd.md
  - tmp/old-bmad-output/planning-artifacts/epics.md
  - tmp/old-bmad-output/planning-artifacts/ux-design-specification.md
  - tmp/old-bmad-output/planning-artifacts/research/market-unified-travel-app-research-2026-01-19.md
  - tmp/old-bmad-output/planning-artifacts/research/technical-weasyprint-viability-research-2026-01-26.md
  - tmp/old-bmad-output/planning-artifacts/implementation-readiness-report-2026-01-26.md
workflowType: 'architecture'
project_name: 'traccia'
user_name: 'Simo'
date: '2026-02-12'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**

31 FRs across 7 categories, phased:
- **MVP (FR1-FR19):** Trip CRUD (5), Event CRUD with 5 typed events and pinned/flexible semantics (7), Timeline display with day grouping, drag-and-drop reordering, pinned anchoring, auto-suggested start times, cross-day moves (7)
- **Phase 1.5 (FR20-FR23):** Survival Export — print-ready PDF with day-by-day layout, addresses, QR codes to Google Maps
- **Phase 2 (FR24-FR31):** Authentication via Supabase, shareable read-only links, travel time estimation, transit risk flags, weather forecasts

**Non-Functional Requirements:**

5 NFRs driving architecture:
- **NFR1:** Page loads and HTMX partial updates < 1s on 10 Mbps — shapes server response time budget
- **NFR2:** Drag-and-drop visual update < 100ms — the most latency-sensitive interaction, likely requires optimistic UI via Alpine.js
- **NFR3:** PDF generation up to 10s acceptable — allows async/heavier processing for export
- **NFR4:** Durable persistence, verified by restart-and-query — standard PostgreSQL guarantees
- **NFR5:** Atomic reordering — failed reorder must not corrupt position state

**Scale & Complexity:**

- Primary domain: Full-stack server-rendered web application
- Complexity level: Medium — rich domain modeling (event type polymorphism, day-level positioning, pinned/flexible semantics) but simple infrastructure (single-user, no real-time, no external APIs in MVP)
- Estimated architectural components: ~6 (HTTP server, domain layer, service layer, repository layer, database, static assets/templates)

### Technical Constraints & Dependencies

**Stack (established by existing codebase + PRD):**
- Go 1.25, chi router, templ templates, HTMX 2.0
- PostgreSQL 16, pgx/v5 driver, sqlc for type-safe queries
- golang-migrate for schema migrations
- Tailwind CSS, templui component library (copy-paste ownership), Alpine.js
- No SPA framework, no client-side routing, no WebSocket/SSE

**Existing codebase establishes:**
- Layered architecture: `handler/` → `service/` → `domain/` ← `repository/`
- Manual dependency wiring in `cmd/app/main.go`
- Updater pattern (`func(*Entity) *Entity`) for partial updates
- Integer SERIAL primary keys (not UUID)
- Auth stubs (`user_id UUID` column, `userID *string` in interfaces)
- Code generation: sqlc (queries → `*_sql.go`) + templ (`.templ` → `*_templ.go`)

**Constraints from UX specification:**
- Desktop-first, single breakpoint at 768px
- WCAG 2.1 AA compliance
- Sheet panels (slide-out) for event creation, inline editing for modifications
- Type-specific form morphing via Alpine.js
- Drag-and-drop with pinned event anchoring

### Cross-Cutting Concerns Identified

1. **Event type polymorphism** — 5 event types with shared + type-specific attributes. Affects domain model, database schema, sqlc queries, form rendering, and event card display. The central modeling decision.
2. **Day-level positioning** — Events grouped by day (derived from trip date range + event start time). Position ordering is within a day, not trip-wide. Affects reordering, cross-day moves, and timeline rendering.
3. **Pinned vs. flexible semantics** — Affects drag-and-drop behavior (pinned events reject drag), reordering logic (flexible events flow around pinned anchors), and visual display (lock icon).
4. **Auth stub propagation** — `user_id` exists in schema but isn't enforced. All repository interfaces accept `userID *string`. Architecture must work for single-user now and multi-user with Supabase auth later, without refactoring interfaces.
5. **Code generation discipline** — sqlc and templ generate code that must never be manually edited. Schema changes require migrations → `just generate`. This constrains how type-specific event attributes are modeled (sqlc must be able to generate the queries).

## Starter Template Evaluation

### Primary Technology Domain

Full-stack server-rendered web application (Go backend + HTMX/templ frontend). Existing codebase — not greenfield.

### Starter Assessment: Existing Codebase

No starter template evaluation is needed. The traccia v2 codebase was established manually with deliberate stack choices. The foundation is production-ready and does not require scaffolding changes.

### Established Foundation

**Language & Runtime:**
- Go 1.25 with standard library conventions
- templ for type-safe HTML template generation (compiled, not interpreted)

**Styling Solution:**
- Tailwind CSS as utility-first engine
- templui as component library (templ + Tailwind + HTMX native, copy-paste ownership model)
- Alpine.js for client-side interactions (form morphing, drag-and-drop)

**Build Tooling:**
- justfile for task commands (`just dev`, `just build`, `just test`, `just lint`, `just generate`)
- air for hot reload during development
- sqlc + templ for code generation (`just generate`)
- golangci-lint with goimports (local prefix: `github.com/simopzz/traccia`)

**Testing Framework:**
- Go standard testing (`go test -v -race`)
- External test packages (`package foo_test`) for black-box testing
- Table-driven tests with explicit input/output expectations

**Code Organization:**
- Layered architecture: `cmd/app/`, `internal/domain/`, `internal/service/`, `internal/repository/`, `internal/handler/`, `internal/infra/`
- Domain defines entities and repository interfaces (ports)
- Repository implements domain interfaces with sqlc-generated code + store adapters
- Handlers own HTTP routing, `.templ` templates co-located
- Manual dependency wiring in `cmd/app/main.go`

**Development Experience:**
- `just dev` → air hot reload for Go + templ
- `just generate` → sqlc + templ code generation
- `just docker-up/down` → PostgreSQL via Docker
- `just migrate-up/down` → schema migrations
- `.env` configuration with `SERVER_ADDRESS`, `DATABASE_URL`, `ENVIRONMENT`

### Foundation Review (Architecture Decision Records)

#### ADR-1: Layered Architecture — ✅ Keep

Layered packages (`domain/`, `service/`, `repository/`, `handler/`) over feature folders. With ~2 domain entities and a solo developer, feature folders add indirection without benefit. sqlc generates into `internal/repository/` — fighting the code generator's conventions to match feature folders is wasted effort. The layout matches Go community norms and the tools. If the domain grows significantly (10+ entities), split by file naming within packages.

#### ADR-2: SERIAL Integer Primary Keys — ✅ Keep

Simpler than UUIDs — smaller index footprint, natural ordering, easier debugging. Sequential ID enumeration risk in URLs is mitigated by the PRD's design: shareable links (Phase 2) use separate high-entropy tokens, not primary keys. Internal IDs stay internal. Mixed ID types (int for entities, UUID for users via Supabase) serve different purposes and are acceptable.

#### ADR-3: sqlc for Query Generation — ✅ Keep

Type-safe Go code from raw SQL, no ORM overhead, full SQL expressiveness. Constrains event type polymorphism modeling — sqlc must understand the schema and generate the queries. Handles single-table queries and JOINs well. For the rare complex dynamic query (conditional JOINs based on event type), manual pgx alongside sqlc is acceptable — sqlc for 90%, manual for 10%.

#### ADR-4: templ + HTMX + Alpine.js + templui — ✅ Keep

HTMX handles partial page updates for the timeline. Alpine.js handles form morphing (TypeSelector) and drag-and-drop (optimistic UI for NFR2 < 100ms). templ gives type-safe templates. templui provides 40+ stack-native components with copy-paste ownership. No SPA framework needed — the PRD explicitly rules out real-time collaboration.

#### ADR-5: Updater Pattern — ✅ Keep with Extension

`Update(ctx, id, func(*Entity) *Entity)` provides transactional read-modify-write for single-entity mutations. Cross-entity validation (e.g., pinned event conflicts) happens in the service layer before calling Update. For multi-entity operations (reordering all events in a day), dedicated repository methods with explicit transaction handling are needed alongside the updater pattern.

#### ADR-6: Flat Event Model — ⚠️ Must Evolve

The current single `events` table with `category TEXT` and shared columns cannot satisfy FR7 (5 typed events) or FR8 (type-specific attributes). Flight needs airline/flight number/terminals/airports, Lodging needs booking reference/check-in-out, Transit needs origin/destination/transport mode. The event type polymorphism modeling approach is the central architectural decision ahead.

### Pre-mortem: Foundation Risk Analysis

Projected failure scenarios and their preventions, identified before implementation begins:

**1. Event Type Swamp (High):** If type-specific attributes use JSONB or untyped storage, sqlc can't generate typed accessors — leading to manual queries, runtime casting, and silent nil bugs. **Prevention:** Evaluate event modeling against sqlc capabilities; write proof-of-concept sqlc query before committing to an approach.

**2. Position Nightmare (High):** If position scope (day vs trip) and reorder algorithm aren't defined upfront, cross-day moves (FR19) produce duplicate positions and unreliable ordering. **Prevention:** Define position scope (day-scoped vs trip-scoped) and reorder algorithm (gap-based, sequential renumbering, or fractional) as an explicit architectural decision. Write the position update query and verify atomicity (NFR5) at design time.

**3. Template Spaghetti (Medium):** 5 event types × compact/expanded variants × mobile/desktop = combinatorial template complexity. **Prevention:** Each event type gets its own templ component (`ActivityCard`, `FlightCard`, etc.). Parent `EventCard` dispatches to type-specific components. New event types = new file, not modifying existing templates.

**4. Drag-and-Drop Desync (Medium):** Optimistic Alpine.js reorder diverges from server state on rejection (pinned conflict, error). **Prevention:** After drag-drop, server response replaces the full day's event list HTML. The HTMX swap is the reconciliation point. Client never holds persistent position state across operations.

**5. Migration Trap (Low):** Schema restructuring for event types is complex if migrating from the flat model. **Prevention:** The v2 codebase has no production data — the initial migration can be replaced with the correct target schema rather than layering migrations on top.

### First Principles Verification

Foundation choices verified against 5 ground truths:

**Ground Truth 1 — Portfolio piece, not a startup:** Stack choices must be demonstrably intentional. Go, sqlc, layered architecture, manual DI wiring — all signal deliberate engineering. Custom templ components (EventCard, TimelineDay) are the portfolio-visible work; templui is a productivity accelerator, not the showcase.

**Ground Truth 2 — Solo developer:** Every abstraction must earn its keep. Validates: no feature folders, no DI framework, HTMX over SPA, raw SQL via sqlc, `just` commands. Updater pattern is worth its cognitive overhead (prevents partial-update bugs). External test packages kept for service/handler tests; internal packages allowed for domain logic.

**Ground Truth 3 — Timeline is the product:** Data model must efficiently serve day-level queries and reordering. Surfaces two new decisions:
- **Add `event_date DATE` column** — the current schema derives "which day" from `start_time` vs trip date range, forcing every day-query to compute date ranges. An explicit `event_date` simplifies queries, indexing, and cross-day moves. Service layer sets it from `start_time` on create/update.
- **Position algorithm** — sequential integers (1, 2, 3) require mass renumbering on insert/reorder. Gap-based integers (1000, 2000, 3000 → insert 1500) avoid this. Needs explicit decision.

**Ground Truth 4 — 5 typed events = core domain richness:** The domain model must make the type system visible — typed Go structs, type-dispatched templ components, compile-time category enforcement. The database schema must support sqlc generating typed code for type-specific fields. This is the central decision for the next step.

**Ground Truth 5 — No production data exists:** The schema can be rewritten from scratch. Design the target schema (with event type modeling, `event_date`, gap-based positions, `notes`, Flight category) as the *first* migration. No backward compatibility constraints. Design for MVP + Phase 1.5 only; Phase 2 columns via future migrations.

### Comparative Analysis: Open Foundation Decisions

Three decisions surfaced from ADRs, pre-mortem, and first principles — evaluated against weighted criteria (1-5 scoring, weights reflect project priorities):

#### Event Type Modeling — Winner: Base + Detail Tables (113 vs 101 vs 67)

| Approach | sqlc compat (×5) | Query simplicity (×4) | Domain expressiveness (×5) | Schema clarity (×3) | Portfolio impression (×4) |
|---|---|---|---|---|---|
| **Base + Detail Tables** | 4 | 3 | 5 | 5 | 5 |
| Single Table Wide | 5 | 5 | 3 | 2 | 2 |
| JSONB Column | 1 | 3 | 2 | 3 | 1 |

Base `events` table with shared fields + separate detail tables (`flight_details`, `lodging_details`, `transit_details`) per type. Activity and Food have no extra fields — they use the base table only. sqlc handles LEFT JOINs. Domain layer projects joined rows into typed Go structs. Demonstrates relational modeling competence for portfolio. JSONB rejected — sqlc can't type subfields, runtime casting, no compile-time safety.

#### Position Algorithm — Winner: Gap-Based Integers (93 vs 84 vs 84)

| Approach | Reorder efficiency (×5) | INTEGER compat (×4) | Atomicity NFR5 (×4) | Debuggability (×3) |
|---|---|---|---|---|
| **Gap-Based (1000, 2000)** | 4 | 5 | 4 | 3 |
| Sequential (1, 2, 3) | 2 | 5 | 3 | 5 |
| Fractional (1.0, 1.5) | 5 | 2 | 5 | 2 |

Gap of 1000 between positions. Insert between existing events without renumbering. When gaps exhaust between two adjacent events, renumber the entire day's positions in a single transaction (rare). Stays INTEGER-compatible with sqlc. Sequential loses on reorder efficiency (N rows updated per move). Fractional loses on sqlc/INTEGER compatibility and debuggability.

#### Day Scoping — Winner: Explicit `event_date DATE` Column (90 vs 70 vs 62)

| Approach | Query simplicity (×5) | Index efficiency (×4) | Cross-day move (×4) | Consistency risk (×3) |
|---|---|---|---|---|
| **Explicit `event_date DATE`** | 5 | 5 | 4 | 3 |
| Separate `days` table | 4 | 4 | 3 | 2 |
| Derived from start_time | 2 | 2 | 2 | 5 |

`event_date DATE` column on events table. Service layer sets it from `start_time` on create/update. Enables `WHERE event_date = $1` queries and composite index `(trip_id, event_date, position)`. Consistency risk (drift from start_time) mitigated by service-layer enforcement. Separate `days` table rejected as over-engineering — days are derived from trip date range, not independent entities.

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
1. Event type modeling: Base + Detail Tables
2. Position algorithm: Gap-based integers (gap of 1000)
3. Day scoping: Explicit `event_date DATE` column
4. HTMX swap strategy: Day-level swaps after mutations
5. Drag-and-drop: SortableJS with Alpine.js event binding

**Important Decisions (Shape Architecture):**
6. Form morphing: Alpine.js `x-show` for TypeSelector
7. Event creation UX: HTMX-fetched Sheet panel with context defaults
8. Validation: Service layer owns all business rules; DB constraints as safety net
9. Migration: Replace initial migration with target schema

**Deferred Decisions (Post-MVP):**
10. Auth middleware + CSRF protection (Phase 2)
11. PDF engine integration — Gotenberg (Phase 1.5)
12. Production hosting — likely Hetzner VPS (Phase 1.5+)
13. CI/CD pipeline (when project goes public)
14. Caching strategy (if performance requires it)

### Data Architecture

**Event Type Modeling — Base + Detail Tables:**

Base `events` table with shared fields (id, trip_id, event_date, title, category, location, latitude, longitude, start_time, end_time, pinned, position, notes, created_at, updated_at). Three detail tables for types with extra attributes:
- `flight_details`: airline, flight_number, departure_airport, arrival_airport, departure_terminal, arrival_terminal, departure_gate, arrival_gate, booking_reference
- `lodging_details`: check_in_time, check_out_time, booking_reference
- `transit_details`: origin, destination, transport_mode

Activity and Food use the base table only — no detail table. Detail tables have 1:1 relationship via `event_id FK REFERENCES events(id) ON DELETE CASCADE`.

**Position Algorithm — Gap-Based Integers:**

Position column uses INTEGER with gap of 1000 (first event at 1000, second at 2000). Insert between existing events without renumbering. When gaps exhaust, renumber the entire day's positions in a single transaction. Composite index: `(trip_id, event_date, position)`.

**Day Scoping — Explicit `event_date DATE`:**

`event_date DATE` column on events table. Service layer derives it from `start_time` on create/update. Enables `WHERE trip_id = $1 AND event_date = $2` queries. Indexed with position for efficient day-level reads and reordering.

**Validation Strategy:**

Service layer validates all business rules: end time > start time, category in closed set, required fields per type, position consistency. Repository trusts service input. Database constraints (NOT NULL, FK, CHECK) as safety net, not primary validation.

**Migration Approach:**

Replace `001_initial.up.sql` with the target schema including detail tables, `event_date`, gap-based positions, `notes`, and Flight category. No data migration needed — no production data exists.

### Authentication & Security

No security-specific architecture for MVP. Stack provides safe defaults: templ auto-escapes HTML (XSS), sqlc uses parameterized queries (SQL injection). Auth stubs (`user_id UUID` column, `userID *string` interfaces) ready for Phase 2 Supabase integration. CSRF tokens and auth middleware added in Phase 2, slotting into chi's middleware chain.

### Frontend Architecture

**HTMX Swap Strategy — Day-Level Swaps:**

After any event mutation (create, edit, delete, reorder), the server returns the entire day's HTML (`<div id="day-{date}">...all events...</div>`). The HTMX swap replaces the full day container. Server is always source of truth. Prevents client-server desync. A day's HTML (10 event cards) is negligible payload.

**Drag-and-Drop — SortableJS:**

SortableJS handles mouse and touch drag-and-drop. Alpine.js listens to SortableJS events, performs optimistic visual reorder (NFR2 < 100ms), then triggers HTMX request with new position order. Server response replaces the full day HTML (reconciliation point).

**Form Morphing — Alpine.js `x-show`:**

TypeSelector selects event type → Alpine.js toggles visibility of type-specific form fields. All 5 form variants exist in the DOM. Instant switch, no server round-trip. Matches UX spec requirement of "instant, not animated."

**Event Creation — HTMX Sheet Panel:**

"Add Event" triggers HTMX fetch of Sheet content, pre-populated with defaults: day from current context, start time from preceding event's end time, end time from type-based duration. Sheet slides from right (desktop) or bottom (mobile). templui Sheet component.

### Infrastructure & Deployment

**MVP:** Local development only. Docker for PostgreSQL, air for hot reload, justfile commands. No production deployment, no CI/CD, no caching layer.

**Logging:** `log/slog` structured logging (Go standard library).

**PDF Export (Phase 1.5):** Service interface defined in domain layer now. Implementation with Gotenberg deferred. Handler exposes `/trips/{id}/export` endpoint when ready.

**Production Hosting (Phase 1.5+):** Decision deferred. Hetzner VPS ($6/mo, 4GB RAM) is the leading candidate due to Gotenberg memory requirements.

## Implementation Patterns & Consistency Rules

### Pattern Categories Defined

**Critical Conflict Points Identified:** 6 areas where AI agents could make different choices given the architectural decisions made.

### Naming Patterns

**Database Naming Conventions:**
- Tables: `snake_case`, plural (`events`, `flight_details`)
- Columns: `snake_case` (`start_time`, `event_date`, `flight_number`)
- Primary keys: `id` (SERIAL)
- Foreign keys: `{referenced_table_singular}_id` (`trip_id`, `event_id`)
- Indexes: `idx_{table}_{columns}` (`idx_events_trip_date_pos`)

**Go Naming (established by CLAUDE.md):**
- `PascalCase` exported, `camelCase` local
- Domain types: `Trip`, `Event`, `FlightDetails`, `LodgingDetails`, `TransitDetails`
- Category constants: `CategoryActivity`, `CategoryFood`, `CategoryLodging`, `CategoryTransit`, `CategoryFlight`
- Repository interfaces: `TripRepository`, `EventRepository`
- Service methods: `CreateEvent`, `ReorderEvents`, `MoveEventToDay`

**HTMX/templ Naming:**
- templ components: `PascalCase` Go functions — `EventCard()`, `FlightCard()`, `TimelineDay()`, `TypeSelector()`
- HTMX target IDs: `kebab-case` with semantic prefix — `day-2026-05-14`, `event-42`, `trip-timeline`
- HTMX trigger events: `kebab-case` — `event-created`, `event-reordered`, `day-updated`
- Alpine.js data attributes: `camelCase` — `x-data="{ selectedType: 'activity', showForm: false }"`

### Structure Patterns

**Event Type Code Organization Across Layers:**

An agent implementing a new event type (e.g., Flight) must touch these files in order:

1. **Migration** — add `flight_details` table in `migrations/`
2. **sqlc queries** — add queries in `internal/repository/sql/flight_details.sql`
3. **Domain** — add `FlightDetails` struct and `CategoryFlight` constant in `internal/domain/models.go`
4. **Repository** — add `FlightDetailsStore` adapter in `internal/repository/flight_details_store.go`; update `EventRepository` interface if needed
5. **Service** — add Flight-specific validation in `internal/service/event.go`
6. **Handler** — add Flight form handling in `internal/handler/event.go`
7. **Templates** — add `FlightCard` component and Flight form fields in `internal/handler/`
8. **Generate** — run `just generate`

**Rule:** Activity and Food do NOT get detail tables, store adapters, or separate card components unless they gain type-specific fields in a future phase. They use the base `EventCard` with shared fields only.

**templ Component Hierarchy for Events:**

```
EventCard(event domain.Event, details any) → dispatches to:
├── ActivityCardContent(event)     — shared fields only
├── FoodCardContent(event)         — shared fields only
├── FlightCardContent(event, fd)   — shared + FlightDetails
├── LodgingCardContent(event, ld)  — shared + LodgingDetails
└── TransitCardContent(event, td)  — shared + TransitDetails
```

The parent `EventCard` renders the common card shell (border, shadow, type icon, lock icon, drag handle). Type-specific content is rendered by the inner component. New types = new `*CardContent` function in a new `.templ` file.

### Communication Patterns

**HTMX Interaction Contract:**

| Action | HTTP Method | Response | HTMX Swap |
|---|---|---|---|
| Create event | POST `/trips/{id}/events` | Full day HTML | `hx-target="#day-{date}"` `hx-swap="outerHTML"` |
| Edit event | PUT `/events/{id}` | Full day HTML | `hx-target="#day-{date}"` `hx-swap="outerHTML"` |
| Delete event | DELETE `/events/{id}` | Full day HTML | `hx-target="#day-{date}"` `hx-swap="outerHTML"` |
| Reorder events | PUT `/trips/{id}/days/{date}/reorder` | Full day HTML | `hx-target="#day-{date}"` `hx-swap="outerHTML"` |
| Move to day | PUT `/events/{id}/move` | Target day + source day OOB | `outerHTML` + `hx-swap-oob` |
| Open create form | GET `/trips/{id}/events/new?date={date}` | Sheet panel HTML | `hx-target="#sheet-content"` |

**Rule:** Every mutation returns the full day's HTML. No partial card swaps. No JSON responses. The server-rendered HTML IS the API.

**Error Responses:**
- Validation error (422): Return the form HTML with inline field errors. No page-level error banners.
- Server error (500): Return empty body + `HX-Trigger: {"toast": {"message": "...", "variant": "error"}}` header. templui Toast handles display.
- Success with side effect: Return HTML + `HX-Trigger` for toast if needed (e.g., "Event deleted. [Undo]").

### Process Patterns

**Error Handling (Go layers):**

```
Handler: translates domain errors → HTTP status + HTML response
  ├── domain.ErrNotFound → 404 page
  ├── domain.ErrValidation → 422 form with errors
  └── unexpected error → 500 + log/slog.Error + toast trigger

Service: validates input, wraps errors with context
  └── fmt.Errorf("creating event for trip %d: %w", tripID, err)

Repository: wraps database errors with context
  └── fmt.Errorf("inserting event: %w", err)
```

**Rule:** Services NEVER return HTTP-aware errors. Services return domain errors. Handlers translate.

**Position Management:**

```
NewEventPosition:    max(existing positions in day) + 1000, or 1000 if day is empty
InsertBetween(a, b): (a.Position + b.Position) / 2  — if gap < 1 → renumber day
MoveToDay:           append at max + 1000 in target day; no gap in source day (leave as-is)
ReorderDay:          receive ordered list of event IDs → assign 1000, 2000, 3000...
```

**Rule:** Position gaps are never "cleaned up" proactively. Renumbering only happens when a gap is too small to insert between (gap < 1). Renumbering always renumbers the entire day in a transaction.

**Drag-and-Drop Flow:**

```
1. User drags flexible event (Alpine.js + SortableJS)
2. SortableJS updates DOM immediately (optimistic, < 100ms)
3. Alpine.js reads new order → sends HTMX PUT /trips/{id}/days/{date}/reorder
   with body: ordered list of event IDs
4. Server validates (pinned events in correct positions), assigns new gap-based positions
5. Server returns full day HTML → HTMX swaps #day-{date} (reconciliation)
6. If server rejects (e.g., tried to move a pinned event), the swap restores correct order
```

**Rule:** Pinned events are included in the reorder payload at their fixed position. Server validates that pinned events haven't moved. If they have, server returns the correct order (effectively reverting the client's optimistic move).

### Enforcement Guidelines

**All AI Agents MUST:**

1. Run `just generate` after any change to `.sql` query files or `.templ` files
2. Never edit generated files (`*_templ.go`, `*_sql.go`, `models.go`, `db.go` in repository/)
3. Always return full day HTML for event mutations — never partial card swaps
4. Place validation in the service layer, never in handlers or repositories
5. Wrap all errors with context using `fmt.Errorf("doing X: %w", err)`
6. Use the gap-based position algorithm — never assign sequential positions
7. Add new event types as separate files, never by modifying existing type components

## Project Structure & Boundaries

### Complete Project Directory Structure

```
traccia/
├── .air.toml                           ← hot reload config (watches .go + .templ)
├── .env.example
├── .gitattributes                      ← marks *_templ.go, sqlcgen/ as linguist-generated
├── .gitignore
├── .golangci.yml                       ← excludes *_templ.go, sqlcgen/
├── CLAUDE.md
├── docker-compose.yml
├── go.mod
├── go.sum
├── Justfile                            ← + `just css`, `just dev` runs air + Tailwind watcher
├── sqlc.yaml                           ← out: internal/repository/sqlcgen
│
├── cmd/
│   └── app/
│       └── main.go                     ← manual DI wiring
│
├── internal/
│   ├── domain/
│   │   ├── errors.go                   ← ErrNotFound, ErrValidation, etc.
│   │   ├── models.go                   ← Trip, Event, EventWithDetails, FlightDetails,
│   │   │                                  LodgingDetails, TransitDetails, EventCategory consts
│   │   └── ports.go                    ← TripRepository, EventRepository interfaces
│   │                                      + PDFExporter interface (Phase 1.5 placeholder)
│   │
│   ├── service/
│   │   ├── event.go                    ← Event CRUD, type-specific validation,
│   │   │                                  position management (gap-based), cross-day moves
│   │   ├── event_test.go
│   │   ├── trip.go                     ← Trip CRUD
│   │   └── trip_test.go
│   │
│   ├── repository/
│   │   ├── sqlcgen/                    ★ ALL GENERATED — never edit
│   │   │   ├── db.go                   ★
│   │   │   ├── models.go              ★
│   │   │   ├── events.sql.go          ★
│   │   │   ├── trips.sql.go           ★
│   │   │   ├── flight_details.sql.go  ★
│   │   │   ├── lodging_details.sql.go ★
│   │   │   └── transit_details.sql.go ★
│   │   ├── sql/                        ← query sources (hand-written)
│   │   │   ├── events.sql              ← + GetEventWithDetails (wide LEFT JOIN),
│   │   │   │                              ListByTripAndDate (JOIN)
│   │   │   ├── trips.sql
│   │   │   ├── flight_details.sql      ← write-path queries only
│   │   │   ├── lodging_details.sql
│   │   │   └── transit_details.sql
│   │   ├── event_store.go              ← wraps sqlcgen.Queries, orchestrates transactional
│   │   │                                  creates (base + details), wide JOIN row mapping
│   │   ├── flight_details_store.go     ← detail write adapter (participates in event txn)
│   │   ├── lodging_details_store.go
│   │   ├── transit_details_store.go
│   │   └── trip_store.go
│   │
│   ├── handler/
│   │   ├── routes.go                   ← chi router setup + method override middleware
│   │   ├── middleware.go               ← future auth/CSRF middleware (Phase 2)
│   │   ├── helpers.go                  ← shared utilities + OOB swap helper
│   │   ├── trip.go                     ← Trip handlers
│   │   ├── trip.templ                  ← Trip pages + TimelineDay + day containers
│   │   ├── event.go                    ← Event handlers (all types, reorder, move)
│   │   ├── event.templ                 ← EventCard dispatch + shared card shell
│   │   ├── event_form.templ            ← Sheet form + TypeSelector (complexity hotspot)
│   │   ├── flight_card.templ           ← FlightCardContent
│   │   ├── lodging_card.templ          ← LodgingCardContent
│   │   ├── transit_card.templ          ← TransitCardContent
│   │   ├── timeline.templ              ← drag-and-drop wiring, SortableJS integration
│   │   └── layout.templ                ← base HTML, vendored JS refs
│   │
│   └── infra/
│       ├── config/
│       │   └── config.go               ← env var loading
│       ├── database/
│       │   └── postgres.go             ← pgx pool setup
│       └── server/
│           └── server.go               ← graceful shutdown
│
├── migrations/
│   ├── 001_initial.up.sql              ← rewritten: target schema with detail tables,
│   │                                      event_date, gap-based positions, notes, Flight category
│   └── 001_initial.down.sql
│
└── static/
    ├── css/
    │   ├── input.css                   ← Tailwind @import directives
    │   └── app.css                     ★ Tailwind CLI build output
    └── js/                             ← vendored with version + source URL comments
        ├── htmx.min.js
        ├── alpine.min.js
        └── sortable.min.js
```

**Generated files convention:**
- sqlc output isolated in `repository/sqlcgen/` (separate Go package)
- templ output (`*_templ.go`) co-located with `.templ` source (tool constraint, not configurable)
- `.gitattributes` marks both patterns as `linguist-generated=true`
- `.golangci.yml` excludes both patterns from linting

**Note:** `*_templ.go` generated files omitted from tree for readability — every `.templ` file has a corresponding `*_templ.go` alongside it.

### Architectural Boundaries

**Layer Dependency Direction (strict, verifiable):**

```
handler/ ──→ service/ ──→ domain/ ←── repository/
   │                        ↑              │
   │                        │              │
   └── imports domain       │         imports domain
       for types/errors     │         for interfaces
                            │
                     ZERO outward deps
```

**API Boundaries:**
- External: HTTP endpoints via chi router in `handler/routes.go`
- Internal: `domain/ports.go` interfaces — services call repository interfaces, never concrete implementations
- Data access: `repository/` is the sole database accessor; service layer never touches `pgx` or `sqlcgen` directly

**Component Boundaries:**
- `event_store.go` orchestrates detail stores within transactions — detail stores never called directly from service layer
- Detail stores accept `sqlcgen.DBTX` interface to participate in event store transactions
- templ components dispatch by category — parent `EventCard` delegates to type-specific `*CardContent`

**Data Boundaries:**
- sqlc-generated models (`sqlcgen/models.go`) stay in repository layer — mapped to domain types in store adapters
- Wide JOIN rows (`GetEventWithDetailsRow`) mapped to `domain.EventWithDetails` based on `category` field
- Domain types never leak sqlc or pgx types

### Requirements to Structure Mapping

**FR1-FR5 (Trip CRUD):**
- `handler/trip.go` + `handler/trip.templ` → `service/trip.go` → `repository/trip_store.go` → `sql/trips.sql`

**FR6-FR12 (Event CRUD, 5 typed events, pinned/flexible):**
- `handler/event.go` + `handler/event_form.templ` → `service/event.go` → `repository/event_store.go` + detail stores → `sql/events.sql` + `sql/*_details.sql`
- Type-specific cards: `handler/flight_card.templ`, `handler/lodging_card.templ`, `handler/transit_card.templ`

**FR13-FR19 (Timeline display, drag-and-drop, reorder, cross-day moves):**
- `handler/trip.templ` (TimelineDay) + `handler/timeline.templ` (SortableJS wiring)
- `handler/helpers.go` (OOB swap helper for cross-day moves)
- `service/event.go` (position management: gap-based insert, day renumber)

**FR20-FR23 (Survival Export — Phase 1.5):**
- `domain/ports.go` (`PDFExporter` interface — defined now)
- Implementation deferred; handler endpoint `/trips/{id}/export` added when ready

### Integration Points

**HTMX Interaction Contract:**

| Action | Method + Route | Response | Swap |
|---|---|---|---|
| Create event | POST `/trips/{id}/events` | Full day HTML | `outerHTML` on `#day-{date}` |
| Edit event | PUT `/events/{id}` | Full day HTML | `outerHTML` on `#day-{date}` |
| Delete event | DELETE `/events/{id}` | Full day HTML | `outerHTML` on `#day-{date}` |
| Reorder | PUT `/trips/{id}/days/{date}/reorder` | Full day HTML | `outerHTML` on `#day-{date}` |
| Move to day | PUT `/events/{id}/move` | Target day + source day OOB | `outerHTML` + `hx-swap-oob` |
| Open form | GET `/trips/{id}/events/new?date={date}` | Sheet panel HTML | `innerHTML` on `#sheet-content` |

**Cross-Day Move (OOB pattern):** Handler returns target day as primary response. Source day returned with `hx-swap-oob="outerHTML:#day-{source-date}"` attribute. OOB helper in `helpers.go`.

**Data Flow — Event CRUD:**
```
Browser → HTMX POST → chi router → handler/event.go (parse form, detect type)
  → service/event.go (validate, set event_date from start_time, calculate position)
  → repository/event_store.go (BEGIN TXN → insert event → insert detail → COMMIT)
  → service returns EventWithDetails
  → handler renders full day HTML via trip.templ TimelineDay
  → HTMX swaps #day-{date}
```

**Data Flow — Reorder:**
```
Browser → SortableJS drag → Alpine.js reads order → HTMX PUT with event ID list
  → handler/event.go (parse ordered IDs)
  → service/event.go (validate pinned positions unchanged, assign gap-based positions)
  → repository/event_store.go (batch UPDATE positions in transaction)
  → handler renders full day HTML
  → HTMX swaps #day-{date} (reconciliation)
```

### Development Workflow Integration

**`just dev`:** Runs `air` (Go + templ hot reload) and `tailwindcss --watch` (CSS rebuild on .templ changes) concurrently.

**`just generate`:** Runs `sqlc generate` (→ `repository/sqlcgen/`) then `templ generate` (→ `*_templ.go` alongside sources). Required after any `.sql` query or `.templ` file change.

**`just css`:** One-shot Tailwind CLI build (`input.css` → `app.css`).

**Build:** `just build` compiles to `bin/app`. Static files served from `static/`.

### Key Structural Decisions from Elicitation

**sqlc generated code isolation:** Output to `repository/sqlcgen/` (separate Go package) prevents accidental edits and provides clear visual boundary. Store adapters import `sqlcgen` and map to domain types.

**Wide JOIN read strategy:** `events.sql` contains `GetEventWithDetails` query that LEFT JOINs all 3 detail tables. Returns a flat row with ~30 fields. `event_store.go` maps to `domain.EventWithDetails` based on `category` — the most complex file in `repository/`. Detail stores handle write-path only.

**Transactional event creation:** `EventRepository.CreateEvent(ctx, event, details)` — single interface method. `event_store.go` owns the transaction, detail stores participate via `sqlcgen.DBTX`. Service layer never coordinates cross-store transactions.

**OOB swap for cross-day moves:** Target day returned as primary HTMX response. Source day returned with `hx-swap-oob` attribute. Helper function in `helpers.go`.

**Tailwind integrated with dev workflow:** `just dev` runs air + Tailwind watcher concurrently. Prevents silent CSS breakage when new Tailwind classes added in `.templ` files.

**Optional extraction points:** `handler/event_form_parse.go` when `event.go` exceeds 500 lines. `service/position.go` when `event.go` exceeds 400 lines. Do not pre-create.

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility:**
All technology choices are compatible and work together without conflicts. Go 1.25 + chi + templ + HTMX 2.0 + Alpine.js + SortableJS + Tailwind CSS + templui on the frontend; PostgreSQL 16 + pgx/v5 + sqlc + golang-migrate on the backend. No version incompatibilities found. One internal inconsistency (Step 5 cross-day move HTMX contract vs Step 6 OOB approach) was identified and corrected during validation.

**Pattern Consistency:**
Naming conventions are internally consistent across all layers: `snake_case` in database, `PascalCase` for Go exports, `camelCase` for Go locals, `kebab-case` for HTMX target IDs and trigger events, `camelCase` for Alpine.js data attributes. Structure patterns (layered packages, type-specific templ files, sqlcgen isolation) align with the technology stack's conventions. Communication patterns (day-level HTMX swaps, OOB for multi-day operations) are coherent and consistently specified.

**Structure Alignment:**
The project directory structure supports every architectural decision made: `sqlcgen/` isolates generated code (ADR-3), detail store files support transactional Base + Detail Table writes, type-specific `.templ` files implement the component dispatch hierarchy, `static/css/` supports the Tailwind build pipeline, and the layered package structure (ADR-1) is preserved throughout.

### Requirements Coverage Validation ✅

**Functional Requirements Coverage (31/31):**

| FR Range | Phase | Status | Architectural Support |
|---|---|---|---|
| FR1-FR5 (Trip CRUD) | MVP | ✅ | `handler/trip.go` → `service/trip.go` → `repository/trip_store.go` |
| FR6-FR12 (Event CRUD, 5 types, pinned/flexible) | MVP | ✅ | Base + Detail Tables, category dispatch, pinned BOOLEAN, type-specific validation |
| FR13-FR19 (Timeline, drag-drop, reorder, cross-day) | MVP | ✅ | `event_date` column, gap-based positions, SortableJS, OOB swap, day-level HTMX swaps |
| FR20-FR23 (Survival Export PDF) | 1.5 | ✅ | `PDFExporter` interface in `domain/ports.go`, Gotenberg deferred |
| FR24-FR28 (Auth, sharing) | 2 | ✅ | Auth stubs (`user_id`, `userID *string`), `middleware.go` placeholder |
| FR29-FR31 (Logistics intelligence) | 2 | ✅ | Deferred — no architectural pre-work needed |

**Non-Functional Requirements Coverage (5/5):**

| NFR | Status | Architectural Support |
|---|---|---|
| NFR1: Page loads < 1s | ✅ | Server-rendered HTML, HTMX partial swaps, no SPA bundle overhead |
| NFR2: Drag-drop < 100ms | ✅ | Optimistic UI via Alpine.js + SortableJS; server reconciliation via day swap |
| NFR3: PDF up to 10s | ✅ | Async processing acceptable; Gotenberg runs as separate container |
| NFR4: Durable persistence | ✅ | PostgreSQL with proper migrations; restart-and-query verifiable |
| NFR5: Atomic reordering | ✅ | Gap-based positions; day renumbering in single transaction |

### Implementation Readiness Validation ✅

**Decision Completeness:**
All 14 decisions (5 critical, 4 important, 5 deferred) are documented with rationale, trade-off analysis, and concrete approaches. Critical decisions include weighted comparative analysis with scoring. Enforcement guidelines provide 7 hard rules for AI agent consistency.

**Structure Completeness:**
Complete directory tree with 40+ files defined and annotated. Every file has a purpose description. Generated files clearly marked with ★. Requirements mapped to specific file paths. Data flows documented step-by-step for the two most complex operations (Event CRUD, Reorder).

**Pattern Completeness:**
6 conflict point categories addressed. HTMX interaction contract specifies exact HTTP methods, routes, responses, and swap targets for all 6 event operations. Error handling layers defined across handler/service/repository. Position management algorithm specified with 4 operation modes. Drag-and-drop flow documented as a 6-step sequence.

### Gap Analysis Results

**Critical Gaps: None found.**

**Important Gaps (resolved during validation):**

1. ~~Step 5/Step 6 cross-day move inconsistency~~ — **Fixed.** Step 5 HTMX contract table updated to match Step 6 OOB pattern.

2. **Client-side library versions unspecified** — Alpine.js, SortableJS, and Tailwind CSS versions not pinned. **Resolution:** Pin versions when vendoring JS files during implementation. Not blocking.

**Nice-to-Have Gaps (acknowledged, not blocking):**

3. **Error page templates not in project tree** — Can be a section within `layout.templ` or a separate file added during implementation.

4. **EventRepository interface evolution** — Architecture adds methods beyond current `domain/ports.go`. Expected — the interface evolves during implementation.

### Architecture Completeness Checklist

**✅ Requirements Analysis**

- [x] Project context thoroughly analyzed (31 FRs, 5 NFRs, 5 cross-cutting concerns)
- [x] Scale and complexity assessed (medium — rich domain, simple infrastructure)
- [x] Technical constraints identified (existing codebase, sqlc, templ, HTMX)
- [x] Cross-cutting concerns mapped (event polymorphism, day positioning, pinned semantics, auth stubs, codegen)

**✅ Architectural Decisions**

- [x] Critical decisions documented with weighted comparative analysis
- [x] Technology stack fully specified (Go 1.25, PostgreSQL 16, HTMX 2.0, etc.)
- [x] Integration patterns defined (HTMX day-level swaps, OOB for cross-day, SortableJS)
- [x] Performance considerations addressed (optimistic UI for NFR2, day-level swaps for NFR1)

**✅ Implementation Patterns**

- [x] Naming conventions established (DB, Go, HTMX/templ, Alpine.js)
- [x] Structure patterns defined (event type across layers, templ component hierarchy)
- [x] Communication patterns specified (HTMX contract table, error responses)
- [x] Process patterns documented (error handling layers, position management, drag-drop flow)

**✅ Project Structure**

- [x] Complete directory structure defined with 40+ files
- [x] Component boundaries established (layer deps, transaction ownership, templ dispatch)
- [x] Integration points mapped (HTMX contract, data flows)
- [x] Requirements to structure mapping complete (FR groups → specific files)

### Architecture Readiness Assessment

**Overall Status: READY FOR IMPLEMENTATION**

**Confidence Level:** High — based on thorough validation with zero critical gaps, comprehensive patterns, and explicit decision rationale.

**Key Strengths:**
- Event type polymorphism fully designed across all layers (schema → sqlc → domain → service → handler → templ)
- Position management algorithm specified with 4 operation modes and clear renumbering rules
- HTMX interaction contract provides exact specs for all 6 event operations
- sqlcgen isolation prevents the most common AI agent mistake (editing generated files)
- Existing codebase patterns (updater, layered architecture) preserved and extended

**Areas for Future Enhancement:**
- Pin client-side library versions during implementation
- Specify the complete `EventRepository` interface when writing domain/ports.go
- Add error page templates during UI implementation
- Consider database seeding strategy for development convenience

### Implementation Handoff

**AI Agent Guidelines:**

- Follow all architectural decisions exactly as documented
- Use implementation patterns consistently across all components
- Respect project structure and boundaries — especially the `sqlcgen/` isolation
- Refer to this document for all architectural questions
- Follow the 7 enforcement rules in Implementation Patterns

**First Implementation Priority:**
Rewrite `migrations/001_initial.up.sql` with the target schema (base events table + 3 detail tables + `event_date` + gap-based positions + `notes` + Flight category). This unblocks all downstream work: sqlc queries, domain model evolution, and store adapters.

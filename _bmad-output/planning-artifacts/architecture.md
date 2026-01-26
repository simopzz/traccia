---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/planning-artifacts/ux-design-specification.md
  - _bmad-output/planning-artifacts/research/market-unified-travel-app-research-2026-01-19.md
  - _bmad-output/planning-artifacts/product-brief-traccia-bmad-test-2026-01-20.md
workflowType: 'architecture'
project_name: 'traccia-bmad-test'
user_name: 'simo'
date: '2026-01-26'
status: 'complete'
---

# Architecture Decision Document

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**
The system requires a robust **Orchestration Engine** to handle:
1.  **Timeline Management:** Linear sequence of events with strict temporal validation (Start < End).
2.  **Rhythm Guardian Logic:** Real-time calculation of transit times (using Haversine/Maps API) to detect "Risk" gaps.
3.  **Survival Export:** A dedicated pipeline to render HTML views into high-fidelity PDFs (using `Gotenberg`).
4.  **Data Persistence:** Storage of trips, events, and locations with support for multi-timezone arithmetic.
5.  **Access Control:** Shareable, high-entropy links for read-only access without authentication.

**Non-Functional Requirements:**
-   **Reliability:** PDF generation must be bulletproof (>99.5% success).
-   **Performance:** Read-only views must be lightweight (SSR) for poor network conditions.
-   **Maintainability:** Strong separation of concerns between the Go backend logic and the HTMX fragment rendering.
-   **Privacy:** Data minimization in exports (no external dependencies).

**Scale & Complexity:**
-   **Primary Domain:** Web App (Travel Tech)
-   **Complexity Level:** Medium (due to PDF infra + Timezone logic)
-   **Estimated Architectural Components:** ~6-8 (API/App Server, DB, PDF Worker, Map Client, Cache, Frontend Assets, etc.)

### Technical Constraints & Dependencies

-   **Language/Framework:** Go (Golang) standard library + Chi + HTMX.
-   **PDF Generation:** Requires a Headless Chrome environment via **Gotenberg**.
-   **External APIs:** Google Maps Platform (Places, Distance Matrix).
-   **Deployment:** Docker Compose (Multi-container).
-   **Frontend:** No heavy client-side framework (React/Vue prohibited); logic stays on server or lightweight Alpine.js.

---

## Starter Template Evaluation

### Primary Technology Domain

**Full-Stack Web Application** (Go Backend + HTMX/Alpine Frontend)

### Starter Options Considered

1.  **Go Blueprint (Selected)**: A robust CLI tool that scaffolds production-ready Go applications with specific support for HTMX, Templ, and Tailwind. It avoids "kitchen sink" bloat while providing essential structure.
2.  **go-templ-htmx-template (HoneySinghDev)**: Good, but less actively maintained and more opinionated on the folder structure.
3.  **GHTT (temidaradev)**: A strong contender, but relies on a specific UI library (PinesUI) that might conflict with our "Swiss/Brutalist" design goals.

### Selected Starter: Go Blueprint

**Rationale for Selection:**
Go Blueprint is the industry-standard CLI for modern Go web apps. It allows us to "compose" our stack (Chi + Postgres + HTMX + Tailwind) rather than cloning a monolithic repo. It explicitly supports the **Templ** library, which is critical for our type-safe HTML rendering requirements. It also includes Docker and Makefile setups out of the box, solving our "Deployment Readiness" NFR.

**Initialization Command:**

```bash
# Install the CLI
go install github.com/melkeydev/go-blueprint@latest

# Create the project with our specific stack
go-blueprint create --name traccia-bmad-test \
  --framework chi \
  --driver postgres \
  --advanced \
  --feature htmx \
  --feature tailwind \
  --feature docker
```

**Architectural Decisions Provided by Starter:**

**Language & Runtime:**
-   **Go 1.23+**: Standard modern Go setup.
-   **Templ**: Pre-configured for type-safe HTML generation (essential for our "Rhythm Guardian" logic visualization).

**Styling Solution:**
-   **Tailwind CSS**: Configured with a standalone CLI (no Node.js dependency required for basic builds).
-   **Structure**: `input.css` -> `output.css` pipeline established in Makefile.

**Build Tooling:**
-   **Air**: Live-reload configured for both Go code and Templ templates.
-   **Makefile**: Standardizes `make run`, `make build`, and `make css`.

**Testing Framework:**
-   **Go Test**: Standard library testing folder structure (`_test.go` files co-located or in `tests/`).

**Code Organization:**
-   **Standard Layout**: `cmd/api`, `internal/server`, `internal/database`.
-   **Web Assets**: Dedicated `web/` directory for Templ components and static assets.

**Development Experience:**
-   **Docker Compose**: Ready-to-go `compose.yml` for the Postgres database.
-   **HTMX Integration**: `htmx.min.js` included and served correctly.

---

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
1.  **Authentication Strategy:** Managed (Supabase).
2.  **PDF Engine:** Gotenberg (Docker Container).
3.  **Hosting Infrastructure:** Hetzner VPS (Raw Linux/Docker).

### Data Architecture

*   **Database:** **Postgres** (via Starter).
*   **Auth Provider:** **Supabase Auth**.
    *   *Rationale:* Handles "Sign in with Google" out of the box; generous free tier; integrates well with Postgres.

### Authentication & Security

*   **Strategy:** **Managed Auth (Supabase)**.
*   **Client:** Official Supabase Go Client.
*   **Middleware:** JWT Validation middleware in Go to protect private routes.

### API & Communication Patterns

*   **Internal API:** **Go Interfaces**. The `RhythmService` and `TimelineService` communicate via direct Go method calls, not HTTP.
*   **External API:** **Google Maps (Official Go Client)**.
*   **PDF Service:** **HTTP to Gotenberg**. The app sends a POST request with HTML to the Gotenberg container running on the private Docker network.

### Frontend Architecture

*   **Core:** **HTMX + Templ**. Server-driven UI.
*   **State:** **Alpine.js**. "Local-First" pattern for transient state (e.g., drag-and-drop ordering) before syncing to server.
*   **CSS:** **Tailwind**.

### Infrastructure & Deployment

*   **Hosting:** **Hetzner VPS ($6/mo - 4GB RAM)**.
    *   *Rationale:* Gotenberg requires 1GB+ RAM to run reliably (Chromium). PaaS free tiers (Railway/Render) often cap at 512MB, causing OOM crashes. A raw VPS provides the necessary vertical scaling cheaply.
*   **Orchestration:** **Docker Compose** (managed via Coolify or manual `docker compose up -d`).
    *   *Stack:* App Container + Postgres Container + Gotenberg Container.

---

## Implementation Patterns & Consistency Rules

### Pattern Categories Defined

**Critical Conflict Points Identified:**
3 areas where AI agents could make different choices: Naming, Structure, HTMX Patterns.

### Naming Patterns

**Database Naming Conventions:**
*   **Tables:** `snake_case`, **Plural** (e.g., `users`, `events`).
*   **Columns:** `snake_case` (e.g., `created_at`).
*   **Primary Keys:** `id` (UUIDv4).

**API/Struct Naming Conventions:**
*   **Go Structs:** `CamelCase`.
*   **JSON Tags:** `camelCase` (e.g., `json:"userId"`). *Critical for Alpine.js compatibility.*

### Structure Patterns

**Feature Folders (Domain-Driven):**
Code is organized by **Domain Feature**, not by technical layer.
*   *Good:* `internal/features/timeline/` (contains handler, service, models, templ).
*   *Bad:* `internal/handlers/`, `internal/models/`.

### Communication Patterns

**HTMX Interaction Pattern:**
*   **Success:** Return HTML Fragment (200 OK).
*   **Validation Error:** Return HTML Form with inline errors (422 Unprocessable Entity).
*   **System Error:** Return empty body + `HX-Trigger: {"toast": "Error message"}` header.

**State Management Pattern:**
*   **"Server as Source of Truth"**: Alpine data is initialized via `x-data='JSON_FROM_GO'`. We do not fetch JSON APIs; we fetch HTML fragments that update the state.

---

## Project Structure & Boundaries

### Complete Project Directory Structure

```bash
traccia-bmad-test/
├── cmd/
│   └── api/
│       └── main.go              # Entry point (initializes Chi, DB, Config)
├── internal/
│   ├── config/                  # Envs & Configuration
│   │   └── config.go
│   ├── database/                # DB Connection & Global Queries
│   │   └── database.go
│   ├── middleware/              # Auth, Logging, CORS
│   │   └── auth_middleware.go
│   └── features/                # DOMAIN-DRIVEN FEATURE FOLDERS
│       ├── timeline/
│       │   ├── handler.go       # HTTP Handlers for Timeline
│       │   ├── service.go       # Business Logic (Reordering, Gaps)
│       │   ├── models.go        # DB Structs
│       │   ├── view.templ       # Main Timeline UI
│       │   └── components.templ # Event Cards, Gap Fillers
│       ├── rhythm/
│       │   ├── service.go       # "Guardian" Logic (Haversine/Maps)
│       │   └── types.go
│       ├── export/
│       │   ├── handler.go       # /export/pdf endpoint
│       │   ├── service.go       # Gotenberg Client Logic
│       │   └── print.templ      # PDF-specific layout
│       └── auth/
│           ├── handler.go       # Login/Callback handlers
│           └── service.go       # Supabase Client wrapper
├── web/
│   ├── assets/                  # Static Files
│   │   ├── css/
│   │   │   ├── input.css        # Tailwind Source
│   │   │   └── output.css       # Generated CSS
│   │   └── js/
│   │       ├── htmx.min.js
│   │       └── alpine.js
│   └── layouts/
│       └── base.templ           # Global HTML Shell
├── migrations/                  # SQL Migrations (Postgres)
├── tests/                       # Integration/E2E Tests
├── compose.yml                  # Docker Compose (App + DB + Gotenberg)
├── Dockerfile                   # Go App
├── Makefile                     # Build commands
└── go.mod
```

### Architectural Boundaries

**API Boundaries:**
*   **Public Web:** Standard HTTP/HTML (HTMX) served from `internal/features/*/handler.go`.
*   **PDF Service:** `ExportService` communicates with Gotenberg via HTTP (internal Docker network).

**Requirements to Structure Mapping:**
*   **Epic: Timeline Orchestration** -> `internal/features/timeline/`
*   **Epic: Rhythm Guardian** -> `internal/features/rhythm/`
*   **Epic: Survival Export** -> `internal/features/export/`
*   **Auth System** -> `internal/features/auth/` + `internal/middleware/`

---

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility:**
The decision to use **Hetzner VPS (4GB RAM)** is the linchpin that makes **Gotenberg** viable. If we had chosen a constrained PaaS, Gotenberg would have failed. This infrastructure choice enables the application architecture.

**Structure Alignment:**
The **Feature Folder** structure aligns perfectly with **Domain-Driven Design**. The `timeline` feature encapsulates both the UI (`view.templ`) and the logic (`service.go`), minimizing cognitive load for both human and AI developers.

### Requirements Coverage Validation ✅

**Functional Requirements Coverage:**
*   **Timeline Orchestration:** Supported by `internal/features/timeline` and the Postgres schema.
*   **Rhythm Guardian:** Supported by `internal/features/rhythm` and the Google Maps API client.
*   **Survival Export:** Supported by `Gotenberg` container and `internal/features/export`.

---

## Architecture Completion Summary

### Final Architecture Deliverables

**📋 Complete Architecture Document**
*   **Tech Stack:** Go 1.23, Chi, HTMX, Alpine.js, Tailwind, Postgres, Gotenberg.
*   **Infrastructure:** Hetzner VPS via Coolify/Docker Compose.
*   **Key Patterns:** "Local-First" Alpine state, "Feature Folders" for Go/Templ co-location.

**🏗️ Implementation Ready Foundation**
*   **Decision Count:** 8 Critical Decisions.
*   **Structure:** Full file tree defined.
*   **Validation:** Confirmed that 4GB RAM VPS solves the PDF memory risk.

### Implementation Handoff

**First Implementation Priority:**
Run the Go Blueprint CLI command to scaffold the project, then add the `docker-compose.yml` for Gotenberg.

```bash
go-blueprint create --name traccia-bmad-test --framework chi --driver postgres --advanced --feature htmx --feature tailwind --feature docker
```

**Architecture Status:** READY FOR IMPLEMENTATION ✅

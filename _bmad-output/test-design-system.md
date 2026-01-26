# System-Level Test Design

## Testability Assessment

- **Controllability**: **PASS**. The chosen stack (Go, Chi, Postgres) is highly testable. The architecture explicitly mentions Service layer injection, which facilitates mocking of external dependencies like Google Maps and Gotenberg. Database seeding is supported by the "Fixture Architecture" in the KB.
- **Observability**: **PASS**. Go standard logging is expected. The HTMX architecture means responses are standard HTML, which is easily inspectable in Integration/E2E tests without complex JS state interception.
- **Reliability**: **PASS**. The backend is stateless (except for Postgres persistence), enabling parallel test execution. The "Local-First" Alpine.js state is transient, reducing server-side state complexity.
- **Concern**: Verifying the *visual correctness* of the PDF export in automated tests is non-trivial. Text content can be verified, but layout requires visual regression tools.

## Architecturally Significant Requirements (ASRs)

1.  **PDF Export Robustness (NFR1)**
    -   **Requirement**: > 99.5% success rate, < 15s latency for PDF generation.
    -   **Risk Score**: **9** (Probability: 3 [External Service/Container dependency], Impact: 3 [Core "Survival" Value]).
    -   **Testability Impact**: Requires a running Gotenberg instance in the test environment (CI). Needs robust timeout and fallback testing.

2.  **Rhythm Guardian Logic (FR6/FR7)**
    -   **Requirement**: Real-time calculation of transit times and risk flagging based on geolocation.
    -   **Risk Score**: **6** (Probability: 2 [Complex Math/API], Impact: 3 [User Trust/Safety]).
    -   **Testability Impact**: Requires precise unit testing of the "Time Math" and Haversine logic, including edge cases (timezone crossings, zero duration).

3.  **Shareable Link Entropy (NFR5)**
    -   **Requirement**: High-entropy, unguessable URLs for read-only access.
    -   **Risk Score**: **6** (Probability: 2, Impact: 3 [Privacy/Security]).
    -   **Testability Impact**: Requires statistical unit tests to verify uniqueness and entropy of the generated hashes (Sqids/UUID).

4.  **Mobile Read-Only Performance (NFR3)**
    -   **Requirement**: Time to Interactive < 2s on 3G networks.
    -   **Risk Score**: **4** (Probability: 2, Impact: 2).
    -   **Testability Impact**: Requires performance testing (k6) simulating poor network conditions.

## Test Levels Strategy

Based on the **Web Application** architecture (Go + HTMX):

-   **Unit Tests (60%)**:
    -   **Scope**: Pure business logic (RhythmService, TimelineService), Domain Models, Utility functions (Time calculation, Hash generation).
    -   **Rationale**: Go's strong typing and fast execution make this the most efficient layer for complex logic like the Rhythm Guardian.

-   **Integration Tests (30%)**:
    -   **Scope**: HTTP Handlers, Middleware (Auth), Database Queries (Postgres), Service Integration (Gotenberg client, Maps client).
    -   **Rationale**: Critical to verify the HTMX response fragments and DB interactions. `testcontainers-go` should be used for ephemeral Postgres instances.

-   **E2E Tests (10%)**:
    -   **Scope**: Critical User Journeys (Sarah's "Panic to Peace", David's "Opportunity Fill"), PDF Export flow.
    -   **Rationale**: Playwright is needed to verify the final HTML rendering, client-side Alpine.js interactions (Drag & Drop), and the PDF download experience.

## NFR Testing Approach

-   **Security**:
    -   **Auth**: Integration tests for Supabase JWT verification middleware. Verify 401/403 responses.
    -   **Data**: Unit tests for ID hashing (entropy).
    -   **Tools**: Go `testing`, OWASP ZAP (optional for periodic scans).

-   **Performance**:
    -   **Load**: k6 tests targeting the Shared Link endpoint (high concurrency) and PDF Generation (resource intensive).
    -   **Metrics**: Track Latency (p95) and Error Rate.

-   **Reliability**:
    -   **Fault Injection**: Mock Google Maps API errors/timeouts in Integration tests to verify "Rhythm Guardian" degrades gracefully.
    -   **PDF Fallback**: Mock Gotenberg timeouts to ensure the system handles failures (e.g., returns appropriate error or triggers async fallback).

-   **Maintainability**:
    -   **Code Quality**: Enforce "Feature Folders" structure.
    -   **Observability**: Verify structured logs (Trace IDs) in Integration tests.

## Test Environment Requirements

-   **Local Development**:
    -   Docker Compose: App, Postgres, Gotenberg.
    -   Mocks: Google Maps API (via Interface mocking or WireMock).

-   **CI Pipeline**:
    -   Services: Postgres (Testcontainers or Service), Gotenberg (Service).
    -   Secrets: Supabase keys (test env), Maps API Key (or Mock).

## Testability Concerns (if any)

-   **External API Costs (Google Maps)**: High risk of incurring costs during CI/automated testing.
    -   **Mitigation**: Strict interface-based mocking for `RhythmService`. **BLOCKER** if not implemented in Sprint 0.
-   **PDF Content Verification**: Validating the PDF content in E2E tests is difficult.
    -   **Mitigation**: Use `pdf-parse` in Playwright to verify text. Rely on Visual Regression testing for the *HTML Print View* (source of truth for PDF) rather than the PDF binary itself.

## Recommendations for Sprint 0

1.  **Define Interfaces First**: Ensure `RhythmService` depends on a `MapsProvider` interface, not the concrete Google Maps client, to enable easy mocking.
2.  **Setup Testcontainers**: Configure `testcontainers-go` for Postgres to enable robust integration testing from Day 1.
3.  **Mock PDF Generator**: Create a "Dev Mode" PDF service that returns a dummy PDF to allow UI development without running Gotenberg constantly.

## Project Scoping & Phased Development

### MVP Strategy & Philosophy

**MVP Approach:** **"The Digital Binder" (Problem-Solving MVP)**
*   **Philosophy:** Focus purely on *reliability* and *organization* first. We must prove we are better than a spreadsheet before we try to be smarter than a travel agent.
*   **Resource Requirements:** 1 Full-Stack Dev (Go/HTMX), 1 Designer (CSS/Print Layout focus).

### MVP Feature Set (Phase 1)

**Core User Journeys Supported:**
*   **Sarah (The Anxious Planner):** Manual entry, timeline visualization, risk detection, PDF export.

**Must-Have Capabilities:**
*   **Orchestrator Timeline:** CRUD events, Drag-and-drop reordering, Timezone awareness.
*   **Rhythm Guardian v0.1:** Simple "Time Math" alerts (e.g., "Transit time > Gap time").
*   **Survival Export:** High-fidelity PDF generation with "Taxi Cards" (static data).
*   **Manual Data Entry:** No APIs, no imports. User types in "Hotel", "Address", "Time".

### Post-MVP Features

**Phase 2: Growth ("The Magic Assistant"):**
*   **Opportunity Filler:** Google Places API integration for "Gap Filling."
*   **Mobile Read-Only View:** Shareable links for "Ben" (The Reluctant Companion).
*   **Account/Auth:** User accounts to save trips (MVP can be local-storage or simple token-based).

**Phase 3: Expansion ("The Ecosystem"):**
*   **Collaborative Editing:** Real-time sync for groups.
*   **Calendar Integration:** Google/Apple Cal read access.
*   **Smart Rhythm:** Weather integration and "Energy Scoring."
*   **Email Parsing:** Ingestion of booking confirmations.

### Risk Mitigation Strategy

**Technical Risks:**
*   **PDF Generation Complexity:** Generating pixel-perfect, print-ready PDFs from HTML is notoriously hard.
    *   *Mitigation:* Use a dedicated headless browser service (e.g., Gotenberg or Puppeteer) dockerized within the backend, rather than client-side JS libraries which are flaky.

**Market Risks:**
*   **"Empty Box" Problem:** Without API imports, manual entry is high friction.
    *   *Mitigation:* Provide "Pre-filled Templates" (e.g., "3 Days in Tokyo") so users can see the value of the timeline visualization immediately without typing.

**Resource Risks:**
*   **Solo Dev Burnout:** Scope creep with "nice to haves."
    *   *Mitigation:* Strict adherence to "No External APIs" in Phase 1 (except maybe a Map tile provider). No Google Places, No Weather, No Auth integrations in MVP.

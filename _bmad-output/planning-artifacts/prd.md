---
stepsCompleted: [step-01-init, step-02-discovery, step-03-success, step-04-journeys, step-05-domain, step-06-innovation, step-07-project-type, step-08-scoping, step-09-functional, step-10-nonfunctional, step-11-polish]
inputDocuments:
  - _bmad-output/planning-artifacts/product-brief-traccia-2026-01-20.md
  - _bmad-output/planning-artifacts/research/market-unified-travel-app-research-2026-01-19.md
  - _bmad-output/planning-artifacts/brainstorming.md
documentCounts:
  briefCount: 1
  researchCount: 1
  brainstormingCount: 1
  projectDocsCount: 0
workflowType: 'prd'
classification:
  projectType: web_app
  domain: travel_tech
  complexity: medium
  projectContext: greenfield
---

# Product Requirements Document: traccia

**Author:** simo
**Date:** 2026-01-20
**Version:** 1.0 (Polished)

## 1. Executive Summary

**traccia** is a **Constraint-Based Travel Orchestrator** designed to solve "Stitching Fatigue"—the exhaustion caused by manually connecting logistics (TripIt) with discovery (Wanderlog).

Unlike competitors that focus on *maximizing content* (places to go), traccia optimizes *continuity* (the flow between them). It treats **User Energy** and **Cognitive Load** as finite resources. By fusing intelligent gap-filling algorithms, burnout prevention logic ("Rhythm Guardian"), and military-grade offline redundancies ("Survival Export"), we serve the **"Anxious Planner"** who demands certainty and the **"Pragmatic Maximizer"** who demands efficiency.

**Core Value Proposition:** Move the user from "Optimistic Planning" to "Realistic Execution."

---

## 2. Success Criteria

### User Success
*   **Trust Loop (Retention):** > 40% of users who generate a "Survival Export" return to the app within 7 days of their trip ending. (Proxy for "The system worked").
*   **Gap Fill Acceptance:** > 30% of "Opportunity Filler" suggestions are accepted and added to the itinerary. (Target vs industry avg < 5%).
*   **Burnout Resolution:** > 60% of "Rhythm Guardian" warnings (e.g., "Too Rushed") result in a user modification to the schedule.

### Business Success
*   **Viral Coefficient (K-Factor):** > 1.2 new users referred per existing user via "Shared Itinerary" links or QR codes on printed exports.
*   **Stitching Volume:** Validation of the core problem solution is measured by the volume of external data points (events, bookings) added to a single itinerary.

### Technical Success
*   **Print Reliability:** "Survival Export" view renders correctly on mobile and desktop browsers 99% of the time. Critical for the "Anxious Planner."
*   **Export Generation:** PDF/Static artifact generation must be robust; failure here breaks the core value proposition.

### Measurable Outcomes
*   **Session "Depth":** Average Edit Session < 5 minutes. (Success = efficiency/speed of planning, not time spent in app).
*   **Safety Net Usage:** 50% of beta users generate a "Survival Export" before their trip.

---

## 3. Product Scope & Phased Development

### MVP Strategy: "The Digital Binder"
*   **Philosophy:** Focus purely on *reliability* and *organization* first. Prove we are better than a spreadsheet before trying to be smarter than a travel agent.
*   **Resource Requirements:** 1 Full-Stack Dev (Go/HTMX), 1 Designer (CSS/Print Layout focus).

### Phase 1: MVP (Minimum Viable Product)
*   **Target User:** Sarah (The Anxious Planner).
*   **Core Capabilities:**
    *   **Orchestrator Timeline:** Manual entry, CRUD events, Drag-and-drop reordering, Timezone awareness.
    *   **Rhythm Guardian v0.1:** Simple "Time Math" alerts (e.g., "Transit time > Gap time").
    *   **Survival Export:** High-fidelity PDF generation with "Taxi Cards" (static data).
    *   **Data Entry:** Manual only (No APIs, no imports).
    *   **Access:** Single-player, persistent storage via simple DB.

### Phase 2: Growth ("The Magic Assistant")
*   **Target User:** David (The Pragmatic Maximizer) + Ben (Reluctant Companion).
*   **Capabilities:**
    *   **Opportunity Filler:** Google Places API integration for "Gap Filling."
    *   **Mobile Read-Only View:** Shareable links for companions.
    *   **Account/Auth:** User accounts for cross-device sync.

### Phase 3: Expansion ("The Ecosystem")
*   **Capabilities:**
    *   **Collaborative Editing:** Real-time sync for groups.
    *   **Calendar Integration:** Google/Apple Cal read access.
    *   **Smart Rhythm:** Weather integration and "Energy Scoring."
    *   **Email Parsing:** Ingestion of booking confirmations.

---

## 4. User Journeys

### 1. Sarah: The "From Panic to Peace" Journey (Primary User - Reliability Focus)
**Backstory:** Sarah (34, PM) is organizing a 10-day Japan trip. She fears "The Domino Effect"—one missed train ruining the trip.
*   **Opening:** Sarah stares at a spreadsheet at 11 PM, unsure if she can make a 7 PM dinner after a 6 PM ticket. Panic sets in.
*   **Action:** She enters events into traccia. The **Rhythm Guardian** flashes "Yellow Alert": *"Transit risk: Shinjuku is 45 mins away. You have 15 mins buffer."*
*   **Relief:** She moves dinner to 8 PM. The timeline turns green. She feels validated.
*   **Climax:** Two days pre-trip, she hits "Generate Survival Export." A bilingual PDF with "Taxi Cards" appears. She prints it.
*   **Resolution:** In Tokyo, her phone dies. Her husband uses the printout to show a taxi driver the address. They arrive safely.

### 2. David: The "Serendipity Engineered" Journey (Primary User - Efficiency Focus)
**Backstory:** David (28, Solo) hates "tourist traps" and "wasting time." He wants efficient discovery.
*   **Opening:** It's 2 PM in Berlin, raining. David has a 4-hour gap before dinner. He hates scrolling generic lists.
*   **Action:** He taps **"Fill Opportunity"** on the timeline gap.
*   **Magic:** The system suggests: "Do You Read Me?!" (Bookstore) -> "Five Elephant" (Coffee). Fits strictly within his 4 hours and location.
*   **Resolution:** He follows the plan. It feels serendipitous but was engineered by constraints.

### 3. Alex: The System Monitor (Admin/Dev - Operational Focus)
**Backstory:** Alex runs the platform. He fears API cost spikes.
*   **Action:** He monitors the "Google Places Quota" dashboard.
*   **Control:** He toggles "Aggressive Caching" when a spike occurs, serving cached results to save costs.

### 4. Ben: The Reluctant Companion (Secondary User - Passive)
**Backstory:** Ben is Sarah's husband. He hates planning but needs to know "Where are we going?"
*   **Action:** Sarah sends him a `traccia.app/share/...` link.
*   **Resolution:** He opens it on airport Wi-Fi. It loads instantly (no login). He sees the hotel address and is informed.

---

## 5. Functional Requirements (Capabilities)

### Timeline Orchestration
*   **FR1:** Users can create a trip with a specific destination and date range.
*   **FR2:** Users can manually add events with: Title, Location (Address + Lat/Long), Start Time, End Time, Category.
*   **FR3:** Users can drag-and-drop events to reschedule within the timeline.
*   **FR4:** Users can view the timeline in a linear, single-stream format.
*   **FR5:** The system must calculate durations using UTC deltas to support multi-timezone trips.

### Rhythm Guardian (Risk Detection)
*   **FR6:** The system must calculate geographical distance between events using Haversine formula (Crow-Flies) based on Lat/Long.
*   **FR7:** The system must flag a "Transit Risk" alert if (Time Gap) < (Estimated Travel Time + Buffer).
*   **FR8:** Users can manually override "Travel Time" for a specific gap to clear a flag.
*   **FR9:** The system must visually indicate "Risk Level" (Green/Yellow/Red).

### Survival Export
*   **FR10:** Users can generate a printable "Tactical Field Guide" (PDF).
*   **FR11:** The Export must render "Taxi Cards" for lodging/activity with large, high-contrast address text.
*   **FR12:** The Export must include static QR codes deep-linking to Google Maps (`geo:` or `https://maps.google.com/?q=`).
*   **FR13:** The Export must follow a "Day-by-Day" chronological layout.

### Data Management
*   **FR14:** Users can save itinerary data (persisted to backend DB).
*   **FR15:** Users can "Clear/Reset" a trip.
*   **FR16:** The system must validate input types (End Time > Start Time).

### Access & Sharing (MVP)
*   **FR17:** Users can generate a "Shareable Link" (hashed URL) for read-only access.
*   **FR18:** Read-only views must load without authentication.
*   **FR19:** Read-only views must be responsive on mobile viewports (375px+).

---

## 6. Non-Functional Requirements (Quality Attributes)

### Reliability & Availability
*   **NFR1 (Export Robustness):** PDF Generation success rate > 99.5%. Fallback: Asynchronous "Email me when ready" flow if generation > 15s.
*   **NFR2 (Data Durability):** Trip data persisted to backend DB. Restorable via unique Trip ID/Token.

### Performance
*   **NFR3 (Mobile Load):** Read-only view "Time to Interactive" < 2s on 3G networks.
*   **NFR4 (Export Latency):** Synchronous PDF generation target < 15s.

### Security & Privacy
*   **NFR5 (Link Entropy):** Shareable links/tokens must use high-entropy strings (UUIDv4/16-char) to prevent enumeration.
*   **NFR6 (Data Minimization):** PDF exports must have NO external dependencies (tracking pixels, remote fonts) for privacy and offline rendering.

### Accessibility
*   **NFR7 (Print Legibility):** PDF must adhere to WCAG AA contrast ratios and use min 12pt font for body text.

---

## 7. Domain & Technical Requirements

### Domain Specifics (Travel Tech)
*   **Data Freshness:** "Opportunity Filler" (Phase 2) must validate opening hours against *future* arrival time, not query time.
*   **Staleness Risk:** Cached API data must include "Last Fetched" timestamps and adhere to Google Places TOS caching limits (usually 24h).
*   **Localization:** "Taxi Cards" must render addresses in the destination's local script (e.g., Japanese for Tokyo) to be usable by local drivers.

### Technical Architecture (Web App)
*   **Pattern:** **Server-Side Rendered (SSR)** with **HTMX + Alpine.js**.
*   **Language:** **Go (Golang)** for backend/logic.
*   **Frontend:** HTMX for swapping timeline segments; Alpine.js for lightweight client interactivity. CSS via Tailwind.
*   **PDF Engine:** Headless Chrome (via Go wrapper like `chromedp`) for high-fidelity rendering.
*   **Offline Strategy:** Basic Service Worker to cache current itinerary view for flaky connections.

### Innovation Patterns
*   **Human-Centric Pacing:** Shifting from "Logistics Optimization" to "Energy Optimization."
*   **Constraint-Based Discovery:** Reversing search from "User asks" to "Context suggests."
*   **Anti-Cloud Safety Net:** Explicitly building for offline reliability as a feature, not a fallback.

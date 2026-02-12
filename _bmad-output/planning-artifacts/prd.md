---
stepsCompleted: [step-01-init, step-02-discovery, step-03-success, step-04-journeys, step-05-domain, step-06-innovation, step-07-project-type, step-08-scoping, step-09-functional, step-10-nonfunctional, step-11-polish]
lastEdited: '2026-02-10'
editHistory:
  - date: '2026-02-10'
    changes: 'Post-validation edits: added Executive Summary, fixed FR14 measurability, removed implementation leakage from NFR1/NFR4, added FR31 traceability note'
classification:
  projectType: web_app
  domain: travel_tech
  complexity: medium
  projectContext: greenfield
inputDocuments:
  - _bmad-output/planning-artifacts/product-brief-traccia-2026-02-09.md
  - _bmad-output/planning-artifacts/research/domain-travel-planning-solo-groups-research-2026-02-09.md
  - _bmad-output/brainstorming/brainstorming-session-2026-02-08.md
  - tmp/old-bmad-output/planning-artifacts/product-brief-traccia-2026-01-20.md
  - tmp/old-bmad-output/planning-artifacts/prd.md
  - tmp/old-bmad-output/planning-artifacts/architecture.md
  - tmp/old-bmad-output/planning-artifacts/epics.md
  - tmp/old-bmad-output/planning-artifacts/brainstorming.md
  - tmp/old-bmad-output/planning-artifacts/ux-design-specification.md
  - tmp/old-bmad-output/planning-artifacts/research/market-unified-travel-app-research-2026-01-19.md
  - tmp/old-bmad-output/planning-artifacts/research/technical-weasyprint-viability-research-2026-01-26.md
  - tmp/old-bmad-output/analysis/brainstorming-session-2026-01-19.md
documentCounts:
  briefCount: 2
  researchCount: 3
  brainstormingCount: 3
  projectDocsCount: 0
workflowType: 'prd'
---

# Product Requirements Document - traccia

**Author:** Simo
**Date:** 2026-02-10

## Executive Summary

traccia is a trip planning tool that gives travelers a single, organized timeline for their journey — replacing the fragmented workflow of spreadsheets, Google Maps, and messaging apps. It targets the unserved gap between "I've booked my flights and hotel" and "I know what I'm doing each day" — the planning logistics layer that no existing tool owns.

The product serves two primary archetypes: the **Anxious Planner** who needs certainty that the schedule works, and the **Pragmatic Maximizer** who wants to extract value from every hour without over-planning. Both share the same core need: a single source of truth for what's happening, when, and where.

traccia is not a booking tool, not an AI generator, not a social platform. Users bring their own bookings and intentions; traccia organizes them into a realistic, executable timeline.

**Core design principles:**

- *The plan is disposable, the traveler is sovereign.* The app serves the traveler's intent, not its own logic.
- *The traveler is the sensor, the app is the calculator.* No autonomous tracking; the user drives, the app computes.
- *Buffers are calculated consequences, not explicit entities.* The system derives realistic constraints from event context rather than asking the user to manage them.

This is a portfolio side project built with Go, HTMX, templ, and PostgreSQL — a deliberately uncommon stack emphasizing clean backend architecture and engineering depth beneath a simple surface.

## Success Criteria

### User Success

The product succeeds when a user can plan a real multi-day trip entirely within traccia and find it genuinely better than their spreadsheet. Specifically:

- A user can build a complete trip from first day to last — adding typed events (activities, food, lodging, transit, flights) across multiple days with real addresses and times.
- The day-by-day timeline shows the shape of each day at a glance, and a user can find any detail (e.g., "what's the hotel on day 4?") within a couple of clicks.
- Pinned events stay anchored during reordering; flexible events move intuitively around them.
- The Survival Export produces a PDF a traveler would actually carry — day-by-day layout, addresses, QR codes to Google Maps.

### Business Success

N/A — this is a portfolio side project. There are no revenue, growth, or retention targets. Success is demonstrating engineering competence through a well-scoped, well-built product that solves a real problem.

### Technical Success

The project's primary audience is engineering reviewers evaluating backend quality:

- **Dependency direction integrity:** Zero domain-layer imports from infrastructure. The direction handler → service → domain ← repository is strict and verifiable.
- **Generated code discipline:** Zero manual edits to generated files (`*_templ.go`, `*_sql.go`, `models.go`, `db.go`). Output matches `just generate`.
- **Zero lint violations:** `just lint` passes clean.
- **Error context chain:** Every error return wraps with context (`fmt.Errorf("doing X: %w", err)`). No naked error returns.
- **Domain model expressiveness:** Event types carry distinct attributes and behavior — not a flat struct with a `type` field.
- **Build simplicity:** Clone → `cp .env.example .env` → `just docker-up` → `just migrate-up` → `just dev`. Working app.

### Measurable Outcomes

- End-to-end trip planning works without leaving the app.
- Survival Export renders correctly on desktop and mobile browsers.
- All technical success criteria pass verification.

## Product Scope & Phased Development

### MVP Strategy

**Approach:** Problem-solving MVP — prove that a timeline-based trip planner with typed events is genuinely better than a spreadsheet for organizing a multi-day trip.

**Resource:** Solo developer. Validate by planning a real trip end-to-end in the app.

**Core principle:** Ship the planning experience first. The timeline is the product; everything else builds on it.

### MVP Feature Set (Phase 1)

**Core journeys supported:** Sarah (happy path, edge case), David (gap awareness) — all planning-side interactions.

**Must-have capabilities:**
- Trip CRUD (create, view, edit, delete) with name, destination, date range
- Typed events: Activity, Food, Lodging, Transit, Flight — each with type-specific attributes
- Pinned (fixed-time) vs. flexible (moveable) event semantics
- Day-by-day timeline view showing the shape of each day
- Drill-down to event details
- Drag-and-drop reordering respecting pinned anchors
- Auto-suggested start times based on preceding event's end time

### Phase 1.5 — Survival Export

- Print-ready PDF with day-by-day layout, addresses, QR codes to Google Maps
- Decoupled from MVP to avoid PDF engine infrastructure blocking core development
- PDF engine decision (Gotenberg vs Go-native) deferred until core planning is solid

### Phase 2 — Growth

- Authentication via Supabase (user accounts, multi-device sync)
- Read-only shared links for companions (no login required)
- Rhythm Guardian v1: travel time estimation, transit risk flags, buffer warnings
- Routing API integration (Geoapify)
- Weather-per-stop awareness (Visual Crossing API)

### Phase 3+ — Vision

- Enhanced Survival Export: bilingual Taxi Cards, richer formatting
- Day-level replanning (external disruption or voluntary deviation)
- Holiday/closure checking (Nager.Date API)
- Opportunity Filler: context-aware gap suggestions from POI APIs
- Group coordination (long horizon)

### Risk Mitigation

**Technical risk — PDF generation:** The heaviest infrastructure decision. Previous research rejected WeasyPrint (insufficient CSS Grid support) and selected Gotenberg (1GB+ RAM, Docker). By moving this to Phase 1.5, the MVP is unblocked. The PDF engine decision can be revisited with more context (e.g., Go-native libraries may have matured).

**Scope creep risk:** Two brainstorming sessions produced 18+ feature ideas across 5 themes. The gravitational pull toward logistics intelligence (Rhythm Guardian, weather, routing) during MVP development is the primary risk. Mitigation: MVP is strictly the planning timeline — no API integrations, no external data sources, no intelligence layer.

**No market risk:** Portfolio project with no revenue or growth targets.

## User Journeys

### 1. Sarah: From Spreadsheet Chaos to Confidence (Happy Path)

Sarah is planning a 5-day trip to Tokyo with her partner. It's three weeks out and she has flights booked, a hotel confirmed, and a growing list of "must-see" places scattered across browser tabs, saved Instagram posts, and a half-finished Google Sheet.

She creates a new trip in traccia — "Tokyo, May 12-16" — and starts with the anchors: flights in, flights out, hotel. These go in as pinned events. Immediately the timeline shows five days with the skeleton of the trip visible — she can see which days are empty and which have fixed points.

Over two evenings, she fills in activities day by day. A temple visit on day 2, a food tour on day 3, a day trip to Hakone on day 4. Each event snaps into the timeline with a suggested start time based on when the previous event ends. She drags a museum from the afternoon to the morning and watches the rest of the day adjust. The pinned hotel check-out on the last day stays put — everything else flows around it.

Two days before departure, she hits "Generate Survival Export." A PDF appears: day-by-day layout, every address printed clearly, QR codes linking to Google Maps for each stop. She saves it to her phone and prints a copy.

In Tokyo, her phone battery dies at Shinjuku Station after a long day of photos. She pulls out the printout, shows the taxi driver the hotel address, and arrives without stress. The printout was the safety net that let her actually enjoy the trip.

### 2. Sarah: The Impossible Day (Edge Case)

Same trip, but on day 3 Sarah has been ambitious. She's packed in a morning fish market visit, a sushi-making class, a shrine visit, shopping in Harajuku, and dinner in Shibuya — all between 6 AM and 9 PM.

Looking at the timeline, she sees the day is dense — events are stacked tight with barely any breathing room between them. The shrine visit ends at 2:30 PM and the Harajuku shopping starts at 2:45 PM, but they're across the city. She realizes this won't work.

She drags the shrine visit to day 4 (which is lighter) and immediately day 3 opens up. The remaining events spread more naturally. She doesn't need the app to tell her the schedule was impossible — the timeline's visual density made it obvious at a glance. She fixed it herself in seconds.

### 3. David: Filling the Gaps in Berlin

David has a 3-day weekend in Berlin. He's booked a hotel and has one dinner reservation on Friday night. That's it — three days of open time with a few vague ideas.

He creates the trip and adds his anchors: hotel (pinned), dinner reservation (pinned). The timeline shows three nearly empty days. He starts filling in the things he knows he wants: the East Side Gallery on Saturday morning, a record store he found on Reddit for Sunday afternoon.

Looking at Saturday, he sees a clean gap between the gallery (ends ~noon) and the evening. He adds a lunch spot near the gallery and a bookshop in Kreuzberg for the afternoon. The timeline shows Saturday now has a shape — morning art, midday food, afternoon wandering, evening free.

Sunday still has a morning gap before the record store. He adds a coffee shop and a flea market. Two taps each. The timeline confirms everything fits. He heads to Berlin knowing each day has structure without rigidity — pockets of intention with room to improvise.

### 4. Ben: The Informed Companion

Ben is Sarah's partner on the Tokyo trip. He didn't plan any of it and doesn't want to. He doesn't have the app installed and won't create an account.

The morning of departure, Sarah sends him a link. He opens it on his phone in the airport lounge — no login, no signup modal, just the trip timeline in read-only mode. He can see today's events, tap into details for addresses and times, and swipe through the days.

On day 2, they split up for the afternoon — Sarah goes shopping, Ben wants to find the ramen place from the plan. He opens the link, finds "Day 2," spots the ramen shop, taps the address, and Google Maps opens with directions.

On day 4, Sarah's phone is dead (the Shinjuku incident). Ben pulls up the shared link on his phone to check tomorrow's departure time. It's right there. No asking Sarah, no digging through emails.

Ben never planned anything, never installed anything, never created an account. He just consumed the plan when he needed it.

### Journey Requirements Summary

These journeys reveal the following capability areas:

- **Trip & event CRUD** with typed events and pinned/flexible semantics (all journeys)
- **Day-by-day timeline** that visually communicates density and gaps (Sarah happy path, David, Sarah edge case)
- **Drag-and-drop reordering** with pinned event anchoring (Sarah happy path, Sarah edge case)
- **Auto-suggested start times** for efficient event entry (Sarah happy path, David)
- **Survival Export (PDF)** with addresses and QR codes (Sarah happy path, Ben)
- **Read-only shared links** without authentication (Ben) — Phase 2
- **Visual schedule density** that makes impossible days obvious without explicit alerts (Sarah edge case)

## Web App Specific Requirements

### Technical Architecture

- **Rendering:** Server-side rendered (SSR) multi-page application. Go backend with templ templates, HTMX for partial page updates, Alpine.js for lightweight client-side interactivity (drag-and-drop, modals).
- **Styling:** Tailwind CSS as the utility-first styling engine. [templui](https://templui.io) as a component library for common UI elements (forms, modals, toasts, date pickers) — copy-paste ownership model, stack-native (templ + Tailwind + HTMX), no runtime dependency.
- **No SPA framework.** No React, Vue, or Angular. Logic stays on the server; the browser receives HTML fragments.
- **No real-time requirements.** Single-user planning tool — no WebSocket/SSE infrastructure needed.
- **SEO:** Not applicable. This is a private planning tool, not a content site.

### Browser Support

Modern evergreen browsers only:
- Chrome (latest 2 versions)
- Firefox (latest 2 versions)
- Safari (latest 2 versions)
- Edge (latest 2 versions)

### Responsive Design

- **Desktop:** Primary planning context. Full layout with timeline and event detail views.
- **Mobile (375px+):** Secondary context for quick reference and light editing. Single-column layout.
- **Breakpoint:** 768px as the primary mobile/desktop threshold.

### Accessibility

- WCAG AA contrast ratios for all text.
- Semantic status signals: risk/warning states use icons alongside color (not color alone) to support color-blind users.
- Keyboard navigation: timeline events navigable via Tab, expandable via Enter.
- Touch targets: minimum 44x44px for all interactive elements.
- Support 200% text zoom without breaking timeline layout.

## Functional Requirements

### Trip Management

- **FR1:** Users can create a trip with a name, destination, and date range.
- **FR2:** Users can view a list of their trips.
- **FR3:** Users can edit a trip's name, destination, and date range.
- **FR4:** Users can delete a trip and all its associated events.
- **FR5:** Users can view a trip organized as a day-by-day timeline spanning the trip's date range.

### Event Management

- **FR6:** Users can add an event to a specific day within a trip.
- **FR7:** Users can select an event type from a closed set: Activity, Food, Lodging, Transit, Flight.
- **FR8:** Each event type captures type-specific attributes (e.g., Flight captures airline, flight number, origin/destination airports, terminal/gate; Lodging captures check-in/check-out, booking reference; Transit captures origin, destination, transport mode).
- **FR9:** All event types capture shared attributes: name, location (address), start time, end time, notes.
- **FR10:** Users can edit any attribute of an existing event.
- **FR11:** Users can delete an event from a trip.
- **FR12:** Users can mark an event as pinned (fixed-time, immovable) or flexible (repositionable).

### Timeline & Positioning

- **FR13:** The timeline displays events grouped by day in chronological order.
- **FR14:** The timeline allocates visual weight proportional to event count per day — days with more events are visually distinguishable from days with fewer events.
- **FR15:** Users can drill down from the day view to see full event details.
- **FR16:** Users can reorder flexible events within a day via drag-and-drop.
- **FR17:** Pinned events remain anchored at their position during reordering operations.
- **FR18:** The system suggests a start time for new events based on the preceding event's end time.
- **FR19:** Users can move an event from one day to another within the same trip.

### Survival Export (Phase 1.5)

- **FR20:** Users can generate a print-ready PDF of their trip.
- **FR21:** The PDF displays events in a day-by-day chronological layout.
- **FR22:** Each event in the PDF shows its address.
- **FR23:** Each event with a location includes a QR code linking to Google Maps.

### Authentication (Phase 2)

- **FR24:** Users can create an account and log in.
- **FR25:** Users can access their trips from multiple devices.

### Sharing (Phase 2)

- **FR26:** Users can generate a shareable link for a trip.
- **FR27:** Recipients can view a trip via the shared link without creating an account.
- **FR28:** Shared views are read-only (no editing capabilities).

### Logistics Intelligence (Phase 2)

- **FR29:** The system estimates travel time between consecutive events based on their locations.
- **FR30:** The system flags connections where the time gap between events is shorter than the estimated travel time.
- **FR31:** Users can view weather forecasts per location for their trip dates. (Derived from Product Brief Phase 2 vision and domain research — no user journey currently exercises this capability.)

## Non-Functional Requirements

### Performance

- **NFR1:** Page loads and partial page updates complete in under 1 second on a connection of 10 Mbps or higher.
- **NFR2:** Drag-and-drop reordering completes the visual update in under 100ms — no perceptible delay between drop and re-rendered state.
- **NFR3:** PDF generation (Phase 1.5) may take up to 10 seconds; acceptable given it's a one-time export action.

### Data Integrity

- **NFR4:** Trip and event data is durably persisted. No data loss on normal server restarts, verified by restart-and-query test.
- **NFR5:** Event reordering operations are atomic — a failed reorder does not leave events in an inconsistent state.

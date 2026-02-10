---
stepsCompleted: [1, 2, 3, 4, 5, 6]
inputDocuments:
  - _bmad-output/brainstorming/brainstorming-session-2026-02-08.md
  - _bmad-output/planning-artifacts/research/domain-travel-planning-solo-groups-research-2026-02-09.md
  - old-bmad-output/planning-artifacts/product-brief-traccia-2026-01-20.md
  - old-bmad-output/planning-artifacts/prd.md
  - old-bmad-output/planning-artifacts/architecture.md
  - old-bmad-output/planning-artifacts/brainstorming.md
  - old-bmad-output/planning-artifacts/epics.md
  - old-bmad-output/planning-artifacts/ux-design-specification.md
  - old-bmad-output/analysis/brainstorming-session-2026-01-19.md
  - old-bmad-output/planning-artifacts/research/market-unified-travel-app-research-2026-01-19.md
  - old-bmad-output/planning-artifacts/research/technical-weasyprint-viability-research-2026-01-26.md
  - old-bmad-output/planning-artifacts/implementation-readiness-report-2026-01-26.md
  - old-bmad-output/implementation-artifacts/epic-1-retrospective.md
date: 2026-02-09
author: Simo
project_name: traccia
---

# Product Brief: traccia

<!-- Content will be appended sequentially through collaborative workflow steps -->

## Executive Summary

**traccia** is a trip planning tool that gives travelers a single, organized timeline for their journey — replacing the fragmented workflow of spreadsheets, Google Maps, and messaging apps.

The travel planning space is crowded with booking platforms and AI itinerary generators, but a significant gap persists: no existing tool handles the logistics of *executing* a trip plan. The "connective tissue" between events — realistic buffers, weather awareness, first/last mile planning — remains entirely manual. Traccia targets this gap.

The product serves two archetypes: the **Anxious Planner** who needs certainty that the schedule works, and the **Pragmatic Maximizer** who wants to extract value from every hour without over-planning. Both share the same core need: a single source of truth for what's happening, when, and where.

Traccia is a portfolio side project built with Go, HTMX, templ, and PostgreSQL — a deliberately uncommon stack in this space. The emphasis is on clean backend architecture: a layered domain model with rich event semantics (pinned vs. flexible, typed behaviors), dependency-inverted service boundaries, and type-safe SQL. The product should feel simple on the surface while demonstrating engineering depth underneath.

---

## Core Vision

### Problem Statement

Travelers plan trips across 4-5 disconnected tools: social media for inspiration, booking platforms for transactions, Google Maps for spatial orientation, spreadsheets for scheduling, and messaging apps for group coordination. The planning layer between "I've booked my flights and hotel" and "I know what I'm doing each day" is where fragmentation lives — and no tool owns it.

### Problem Impact

- **Decision fatigue:** 89% of leisure travelers report frustration during planning. Hours spent cross-referencing tools for a single day's logistics.
- **Operational fragility:** Critical details fall through cracks — missing addresses, unrealistic timing, no backup plans. A single missed connection cascades into chaos.
- **Cognitive overhead:** Travelers become manual data processors, maintaining mental models across disconnected systems instead of enjoying the anticipation of the trip.

### Why Existing Solutions Fall Short

- **Wanderlog** offers drag-and-drop itinerary building but treats all time gaps as equal — no awareness of distance, transport mode, or luggage. Users describe it as "clunky and overwhelming."
- **TripIt** excels at organizing existing bookings (email parsing) but offers zero help with planning or discovery.
- **AI generators** (Layla, Mindtrip, Wonderplan) produce generic itineraries from prompts but hallucinate details and lack deep editing tools. They stop at "here's a list of things to do."
- **None** of these tools handle the connective tissue: context-aware buffers, weather-per-stop, first/last mile gaps, or day-level replanning when things break.

### Proposed Solution

Traccia is a **planning logistics layer** — not a booking tool, not an AI generator, not a social platform. Users bring their own bookings and intentions; traccia organizes them into a realistic, executable timeline.

**Core capabilities (MVP):**
- **Trip timeline with typed events:** Activities, food, lodging, and transit as distinct event types with different behaviors and attributes. Events can be pinned (fixed-time anchors like flights) or flexible (moveable within the day).
- **Planning for any scope:** Full-day planning from scratch, or filling a specific time gap with appropriate activities.
- **Basic Survival Export:** A print-ready PDF artifact with addresses, QR codes to Google Maps, and essential trip data — validating the feasibility of offline-ready output.

**Core principles:**
- *"The plan is disposable, the traveler is sovereign."* — The app serves the traveler's intent, not its own logic.
- *"The traveler is the sensor, the app is the calculator."* — No autonomous tracking; the user drives, the app computes.
- *Buffers are calculated consequences, not explicit entities.* — The system derives realistic travel time from event context rather than asking the user to manage buffer objects.

### Key Differentiators

- **Domain modeling depth:** Events aren't flat records — they carry type-specific behavior (transit has origin/destination, lodging spans overnight, activities have duration constraints). Pinned vs. flexible semantics enable realistic scheduling.
- **The unsolved problem:** Context-aware logistics planning is genuinely unserved. AI generation is commoditized; deterministic planning logic is not.
- **Unique tech stack:** Go + HTMX + templ + PostgreSQL with sqlc. No open-source competitor exists with this stack — it's a deliberate portfolio differentiator showcasing backend engineering over frontend frameworks.
- **Architectural clarity:** Layered architecture with dependency inversion, type-safe SQL generation, and clean service boundaries — designed to demonstrate software engineering principles, not just ship features.

## Target Users

### Primary Users

#### 1. Sarah, The "Anxious Planner"
- **Profile:** The designated trip architect for her group or herself. Views travel as a high-stakes investment of time and money. Her enjoyment is directly tied to confidence that "everything is under control."
- **Context:** Solo traveler, couple, or the organizer in a friend group. Under 50, plans 2-5 trips per year (often shorter micro-cations of 4-5 days).
- **The Fear:** "The Domino Effect" — one missed connection spiraling the whole day. Also "The Void" — being stuck in a foreign city with no signal and no idea where the hotel is.
- **Current Workarounds:** Rigid spreadsheets, printed confirmations in a folder, constant Google Maps cross-checking that drains battery and mental energy.
- **What traccia gives her:** A single timeline that replaces the spreadsheet. Events with real addresses and times, organized by day. A Survival Export she can print and put in her pocket — the safety net that lets her relax.
- **Success moment:** She opens the printed PDF in a taxi, shows the driver the address in the local language, and arrives without stress.

#### 2. David, The "Pragmatic Maximizer"
- **Profile:** Solo traveler or lead explorer. Energy-rich but time-poor. Hates "tourist traps" and "dead time." Wants maximum value from every hour without feeling rushed.
- **Context:** Takes frequent short trips (weekends, 4-day breaks). Plans quickly, doesn't want to spend hours organizing. Finds Google Maps "dumb" because it answers "Where is X?" but not "What can I do in this 2-hour gap?"
- **The Friction:** Scrolling "Top 10" lists that don't account for time, location, or what he's already doing that day. No tool helps him fill a specific gap with something that actually fits.
- **Current Workarounds:** Google Maps saved places, mental juggling, asking locals.
- **What traccia gives him:** A timeline where he can see his day's shape at a glance, spot the gaps, and fill them with events that fit the time and location. Planning a day takes minutes, not hours.
- **Success moment:** He adds three things to a free afternoon, sees they fit without rushing, and heads out with confidence.

### Secondary Users

#### 3. Ben, The "Reluctant Companion"
- **Profile:** Sarah's partner, David's travel buddy. Didn't plan the trip and doesn't want to install an app. Just needs to know "What's next?" and "Where are we going?"
- **Interaction:** Consumes the Survival Export (printed PDF) or a read-only shared link. Reduces the friction of asking "What time is the train?" every 20 minutes.
- **What traccia gives him:** Passive access to the plan without effort. He's informed without being involved in planning.

### User Journey

#### Sarah's Journey (From Anxiety to Confidence)
1. **Planning:** She creates a trip, adds her flights and hotel as pinned events, then fills in activities day by day. She sees the full timeline take shape — no more cross-referencing a spreadsheet with Google Maps.
2. **Validation:** She notices a gap she hadn't thought about — the connection from the airport to the hotel. She adds a transit event and sees it fits.
3. **Safety net:** Two days before departure, she generates the Survival Export. A PDF with day-by-day layout, addresses, and QR codes to Google Maps. She saves it to her phone and prints a copy.
4. **In-trip:** Her phone battery dies at Shinjuku Station. She pulls out the printout, shows the taxi driver the address, arrives at the hotel.
5. **Retention:** She uses traccia for the next trip without hesitation.

#### David's Journey (From Dead Time to Full Days)
1. **Planning:** He has a 3-day trip to Berlin. He adds the fixed points (hotel, one dinner reservation) and sees three days of open time.
2. **Gap filling:** He adds a museum, a bookstore, coffee — slotting them into gaps that make geographic and temporal sense. The timeline shows him the shape of each day.
3. **Execution:** On the ground, he checks the timeline, knows exactly where he's headed next, and doesn't waste time deciding.
4. **Retention:** He trusts the tool as his trip organizer — simple, fast, no fluff.

## Success Metrics

### User Success

Success is demonstrated when the product delivers on its core promise without friction:

- **End-to-end planning:** A user can plan a multi-day trip entirely within traccia — creating the trip, adding typed events across days, seeing the shape of each day at a glance, drilling into event details, and generating a Survival Export.
- **Useful export:** The Survival Export produces a PDF that a traveler would actually carry — day-by-day layout, addresses, QR codes to Google Maps. It renders correctly on both desktop and mobile browsers.
- **Snappy CRUD:** Event creation, editing, reordering, and deletion feel instant. This is simple functionality and should behave like it.

### Technical Success

The project's primary audience is engineering reviewers evaluating backend quality:

- **Architectural integrity:** Zero domain-layer dependencies on infrastructure. The dependency direction (handler → service → domain ← repository) is strict and verifiable.
- **Domain modeling quality:** Event types carry distinct semantics (transit, activity, food, lodging). Pinned vs. flexible behavior is modeled at the domain level, not hacked in the UI. The model reflects real travel planning concepts, not just CRUD fields.
- **Clean boundaries:** Each layer has a clear responsibility. Services own business logic and validation. Repositories implement domain interfaces. Handlers translate HTTP to service calls. No layer leaks into another.
- **Code generation discipline:** sqlc for type-safe SQL, templ for type-safe templates. Generated code is never manually edited.

### Business Objectives

N/A — This is a portfolio side project, not a business. There are no revenue, growth, or retention targets. The "business case" is demonstrating engineering competence through a well-scoped, well-built product that solves a real problem.

### Key Performance Indicators

N/A — No formal tracking infrastructure will be built. Success is evaluated by direct use and code review, not dashboards.

## MVP Scope

### Core Features

**Trip Management:**
- Create, view, edit, and delete trips with name, destination, and date range.
- Each trip organizes events into a day-by-day timeline.

**Typed Events (closed set):**
- **Activity** — general events (museums, parks, tours). Location, start/end time, notes.
- **Food** — meals and dining. Location, start/end time, notes.
- **Lodging** — accommodation. Location, check-in/check-out, booking reference, notes. Spans overnight.
- **Transit** — ground transport between locations. Origin, destination, transport mode, start/end time, notes.
- **Flight** — air travel. Airline, flight number, origin/destination airports, terminal/gate, departure/arrival times, booking reference.

All event types support **pinned** (fixed-time, immovable) and **flexible** (repositionable) semantics.

**Timeline & Positioning:**
- Day-by-day view showing the shape of each day at a glance, with drill-down to event details.
- Event reordering via drag-and-drop.
- Smart positioning that respects pinned events — flexible events can be moved around pinned anchors, but pinned events stay put.
- Auto-suggested start times based on the preceding event's end time.

**Basic Survival Export:**
- Print-ready PDF with day-by-day chronological layout.
- Each event shows address and a QR code linking to Google Maps.
- Validates the feasibility of offline-ready output as a core value proposition.

### Out of Scope for MVP

- **Rhythm Guardian** — transit risk detection, buffer warnings, pacing logic. Deferred: requires intersection of routing data, event context, and complex rules. Will evolve incrementally post-MVP.
- **Opportunity Filler** — gap-filling suggestions via external APIs (Google Places, Geoapify). Deferred: requires API integration and suggestion engine.
- **Pre-trip intelligence** — weather per stop, holiday checking, first/last mile awareness. Deferred: requires external API integration.
- **Authentication / user accounts** — auth stubs exist in the codebase but no real auth. Single-user for now.
- **Sharing / collaboration** — read-only links, collaborative editing, Ben's companion view. Post-MVP.
- **Day-level replanning / drift recovery** — deferred: builds on Rhythm Guardian's constraint model.
- **AI features** — not the goal. Traccia's value is deterministic planning logic.
- **Map integration** — no interactive map view in MVP. Events have locations but the UI is timeline-first.

### MVP Success Criteria

The MVP succeeds when:
- A user can plan a multi-day trip end-to-end without leaving the app.
- The day-by-day timeline shows the shape of each day at a glance and supports drill-down to details.
- Typed events carry meaningful, type-specific attributes (not just title + time).
- Pinned events stay anchored during reordering.
- The Survival Export produces a PDF a traveler would actually use.
- The codebase demonstrates clean layered architecture with strict dependency direction.

### Future Vision

**Phase 2 — Logistics Intelligence:**
- Rhythm Guardian v1: travel time estimation between events, transit risk flags, buffer warnings.
- Routing API integration (Geoapify) for realistic travel time calculations.
- Weather-per-stop awareness (Visual Crossing API).

**Phase 3 — Resilience & Export:**
- Enhanced Survival Export: bilingual "Taxi Cards" with addresses in local script, richer formatting.
- Day-level replanning when plans break (external disruption or voluntary deviation).
- Holiday/closure checking (Nager.Date API).

**Phase 4 — Discovery & Sharing:**
- Opportunity Filler: context-aware gap suggestions from POI APIs.
- Read-only shared links for companions.
- Authentication and multi-device sync via Supabase.

**Long-term (if warranted):**
- Group coordination features.
- First/last mile planning.
- Context-aware buffers accounting for luggage, fatigue, and transport mode.

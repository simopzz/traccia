---
stepsCompleted: [step-01-validate-prerequisites, step-02-design-epics, step-03-create-stories]
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/planning-artifacts/architecture.md
  - _bmad-output/planning-artifacts/ux-design-specification.md
---

# traccia-bmad-test - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for traccia-bmad-test, decomposing the requirements from the PRD, UX Design if it exists, and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: Users can create a trip with a specific destination and date range.
FR2: Users can manually add events with: Title, Location (Address + Lat/Long), Start Time, End Time, Category.
FR3: Users can drag-and-drop events to reschedule within the timeline.
FR4: Users can view the timeline in a linear, single-stream format.
FR5: The system must calculate durations using UTC deltas to support multi-timezone trips.
FR6: The system must calculate geographical distance between events using Haversine formula (Crow-Flies) based on Lat/Long.
FR7: The system must flag a "Transit Risk" alert if (Time Gap) < (Estimated Travel Time + Buffer).
FR8: Users can manually override "Travel Time" for a specific gap to clear a flag.
FR9: The system must visually indicate "Risk Level" (Green/Yellow/Red).
FR10: Users can generate a printable "Tactical Field Guide" (PDF).
FR11: The Export must render "Taxi Cards" for lodging/activity with large, high-contrast address text.
FR12: The Export must include static QR codes deep-linking to Google Maps (`geo:` or `https://maps.google.com/?q=`).
FR13: The Export must follow a "Day-by-Day" chronological layout.
FR14: Users can save itinerary data (persisted to backend DB).
FR15: Users can "Clear/Reset" a trip.
FR16: The system must validate input types (End Time > Start Time).
FR17: Users can generate a "Shareable Link" (hashed URL) for read-only access.
FR18: Read-only views must load without authentication.
FR19: Read-only views must be responsive on mobile viewports (375px+).

### NonFunctional Requirements

NFR1: PDF Generation success rate > 99.5%. Fallback: Asynchronous "Email me when ready" flow if generation > 15s.
NFR2: Trip data persisted to backend DB. Restorable via unique Trip ID/Token.
NFR3: Read-only view "Time to Interactive" < 2s on 3G networks.
NFR4: Synchronous PDF generation target < 15s.
NFR5: Shareable links/tokens must use high-entropy strings (UUIDv4/16-char) to prevent enumeration.
NFR6: PDF exports must have NO external dependencies (tracking pixels, remote fonts) for privacy and offline rendering.
NFR7: PDF must adhere to WCAG AA contrast ratios and use min 12pt font for body text.

### Additional Requirements

- **Starter Template (CRITICAL):** Use Go Blueprint CLI to scaffold project (Chi + Postgres + HTMX + Tailwind + Docker). This is the starting point for Epic 1 Story 1.
- **Architecture - Structure:** Use Domain-Driven Feature Folders (`internal/features/timeline`, `rhythm`, `export`, `auth`).
- **Architecture - Auth:** Use Supabase for Authentication (Managed Auth).
- **Architecture - PDF:** Use Gotenberg (Docker Container) for PDF generation.
- **Architecture - Frontend:** Use HTMX + Templ for SSR, Alpine.js for local state.
- **Architecture - Database:** Postgres via Docker Compose. Tables `snake_case` plural; JSON tags `camelCase`.
- **Architecture - API:** Google Maps (Places, Distance Matrix) via Official Go Client.
- **UX - Direction:** "Swiss / Brutalist Field Guide" aesthetic. High contrast, strict grid, distinct "Safety" colors (Green/Amber/Red).
- **UX - Interaction:** "Search -> Time -> Slot" loop.
- **UX - Timeline:** Vertical "Stream of Consciousness". 1 hour = 64px height.
- **UX - Layout:** Desktop Split View (Timeline Fixed + Map Fluid); Mobile Single Column (Timeline default, Map toggle).
- **UX - Typography:** Inter font, Tabular Numerals (`font-variant-numeric: tabular-nums`) for time alignment.
- **UX - Print:** `@media print` stylesheet for "Taxi Cards" (Black/White, no backgrounds).

### FR Coverage Map

FR1: Epic 1 - Trip Creation
FR2: Epic 1 - Event Management
FR3: Epic 1 - Reordering
FR4: Epic 1 - Timeline View
FR5: Epic 1 - Timezone Logic
FR6: Epic 2 - Distance Calculation
FR7: Epic 2 - Risk Detection
FR8: Epic 2 - Manual Override
FR9: Epic 2 - Visual Risk Indicators
FR10: Epic 3 - PDF Generation
FR11: Epic 3 - Taxi Card Layout
FR12: Epic 3 - QR Codes
FR13: Epic 3 - Chronological Layout
FR14: Epic 1 - Data Persistence
FR15: Epic 1 - Clear/Reset
FR16: Epic 1 - Validation
FR17: Epic 4 - Share Links
FR18: Epic 4 - Public Access
FR19: Epic 4 - Mobile Responsiveness

## Epic List

### Epic 1: Trip Core & Timeline Orchestration
Enable users to create a trip, manually build a linear timeline of events, and manage their itinerary data, effectively replacing their spreadsheet.
**FRs covered:** FR1, FR2, FR3, FR4, FR5, FR14, FR15, FR16

### Epic 2: Rhythm Guardian (Logistics Intelligence)
Provide users with real-time feedback on the feasibility of their schedule by calculating transit times and flagging impossible connections.
**FRs covered:** FR6, FR7, FR8, FR9

### Epic 3: Survival Export (Tactical Field Guide)
Empower users to generate high-fidelity, offline-ready PDF artifacts that ensure travel safety and reliability in low-tech environments.
**FRs covered:** FR10, FR11, FR12, FR13

### Epic 4: Shareable Read-Only Access
Allow users to share their itinerary with companions via secure, unguessable links for mobile-friendly viewing without account creation.
**FRs covered:** FR17, FR18, FR19

## Epic 1: Trip Core & Timeline Orchestration

Enable users to create a trip, manually build a linear timeline of events, and manage their itinerary data, effectively replacing their spreadsheet.

### Story 1.1: Project Scaffolding & Database Setup

As a Developer,
I want to initialize the Go Blueprint project with the correct tech stack and folder structure,
So that I have a production-ready foundation for building features.

**Acceptance Criteria:**

**Given** the developer has the Go Blueprint CLI installed
**When** they run the initialization command specified in the Architecture doc
**Then** the project should be created with Chi, Postgres, HTMX, Tailwind, and Docker support
**And** the folder structure should include `internal/features/timeline` and `internal/features/auth`
**And** `make run` should start the server successfully

### Story 1.2: Trip Management (Create/Read/Reset)

As a Planner (Sarah),
I want to create a new Trip with a name, destination, and dates,
So that I have a container to start organizing my itinerary.

**Acceptance Criteria:**

**Given** the user is on the home page
**When** they enter "Japan Trip" and dates and click "Start Planning"
**Then** a new Trip record is created in the database
**And** the user is redirected to the Trip Timeline view (e.g., `/trips/{uuid}`)
**And** the "Clear Trip" button deletes all associated events for that trip

### Story 1.3: Event Creation & Timeline View

As a Planner,
I want to add events with times and locations to my trip and see them in a linear vertical list,
So that I can visualize the flow of my day.

**Acceptance Criteria:**

**Given** a Trip exists
**When** I fill out the "Add Event" form (Title, Address, Start Time, End Time)
**Then** the event is saved to the database with UTC timestamps
**And** the timeline updates via HTMX to show the new event card
**And** the vertical height of the card is proportional to its duration (approx 64px per hour)

### Story 1.4: Drag-and-Drop Reordering

As a Planner,
I want to drag event cards to new timeslots,
So that I can reschedule my day intuitively without manual time entry.

**Acceptance Criteria:**

**Given** multiple events on the timeline
**When** I drag an event card to a new position using the Alpine.js drag handle
**Then** the new time order is sent to the backend via HTMX
**And** the start/end times are updated in the database to reflect the new sequence
**And** the visual order persists after page refresh

## Epic 2: Rhythm Guardian (Logistics Intelligence)

Provide users with real-time feedback on the feasibility of their schedule by calculating transit times and flagging impossible connections.

### Story 2.1: Geolocation & Distance Calculation

As a System,
I want to calculate the distance between consecutive events using their Lat/Long coordinates,
So that I can determine the baseline travel need.

**Acceptance Criteria:**

**Given** two consecutive events with valid Lat/Long coordinates
**When** the timeline is rendered or updated
**Then** the backend calculates the Haversine distance between them
**And** estimates a default travel time (e.g., 50km/h average or Maps API if available)
**And** stores/caches this "Required Transit Time"

### Story 2.2: Transit Risk Logic & Alerts

As a Planner,
I want to see a red warning if my gap between events is smaller than the required travel time,
So that I don't plan impossible transfers.

**Acceptance Criteria:**

**Given** an event ends at 10:00 and the next starts at 10:30 (30m gap)
**And** the calculated travel time is 45 minutes
**When** I view the timeline
**Then** a Red "Risk Alert" component is injected between the events
**And** the visual timeline stream turns red
**And** a text warning says "Impossible Connection: Need ~45m"

### Story 2.3: Manual Override

As a Planner,
I want to manually set the "Travel Time" for a specific gap,
So that I can clear false alarms (e.g., "I'm taking a helicopter").

**Acceptance Criteria:**

**Given** a Risk Alert exists
**When** I click "Override Travel Time" and enter "20 mins"
**Then** the system accepts this new value as the truth
**And** the Red Risk Alert disappears (turns Green)
**And** the override is persisted to the database

## Epic 3: Survival Export (Tactical Field Guide)

Empower users to generate high-fidelity, offline-ready PDF artifacts that ensure travel safety and reliability in low-tech environments.

### Story 3.1: Gotenberg Integration & PDF Service

As a Developer,
I want to set up the PDF generation service using Gotenberg,
So that I can convert HTML views into PDF documents reliably.

**Acceptance Criteria:**

**Given** the Gotenberg container is running via Docker Compose
**When** the backend sends a POST request to Gotenberg with simple HTML
**Then** a valid PDF buffer is returned
**And** the PDF service handles errors gracefully (e.g., timeout > 15s triggers fallback/error)

### Story 3.2: Print Styling & Layout

As a Planner,
I want my PDF export to look like a "Field Guide" with high-contrast text and "Taxi Cards",
So that it is legible in low light and usable by local drivers.

**Acceptance Criteria:**

**Given** a trip with events exists
**When** I request the PDF export
**Then** the output follows the "Day-by-Day" chronological layout
**And** specific CSS `@media print` rules remove background colors and maximize contrast (Black/White)
**And** "Taxi Cards" feature large, bold address text (>12pt)

### Story 3.3: QR Code Generation

As a Traveler,
I want static QR codes on my printout that link to Google Maps,
So that I can quickly load navigation on my phone if I have data.

**Acceptance Criteria:**

**Given** an event has a valid location
**When** the PDF is generated
**Then** a QR code is rendered next to the address
**And** scanning the QR code opens the specific location in Google Maps (`https://maps.google.com/?q=...`)
**And** the QR code is generated server-side (no external JS dependency in PDF)

## Epic 4: Shareable Read-Only Access

Allow users to share their itinerary with companions via secure, unguessable links for mobile-friendly viewing without account creation.

### Story 4.1: Secure Link Generation (Hashids)

As a Planner,
I want to generate a secret link for my trip that is impossible to guess,
So that I can share it without making my trip public to the world.

**Acceptance Criteria:**

**Given** a private trip
**When** I click "Share"
**Then** the system generates a unique hash (e.g., using `sqids` or UUIDv4) associated with the trip ID
**And** returns a URL like `traccia.app/s/{hash}`
**And** the link does not expose the sequential DB ID

### Story 4.2: Mobile Read-Only View

As a Companion (Ben),
I want to open the shared link on my phone and see the plan without logging in,
So that I know where we are going.

**Acceptance Criteria:**

**Given** I have a valid shared link
**When** I open it on a mobile browser
**Then** I see the timeline in a read-only state (no edit buttons)
**And** the layout is optimized for mobile (single column)
**And** no login or signup modal blocks my view
